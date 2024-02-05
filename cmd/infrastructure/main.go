package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecs"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var availabilityZones = []string{"eu-central-1a", "eu-central-1b"}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Setup networking
		vpc, err := ec2.NewVpc(ctx, "grafto-vpc", &ec2.VpcArgs{
			CidrBlock:          pulumi.String("10.0.0.0/16"),
			EnableDnsHostnames: pulumi.Bool(true),
			EnableDnsSupport:   pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		subnets := make(map[string][]*ec2.Subnet, len(availabilityZones))

		var startingSubnetCidrRange = "10.0.0.0/20"

		for i, az := range availabilityZones {
			publicSubnet := fmt.Sprintf("grafto-%s-subnet-%v-%v", ctx.Stack(), "public", i+1)
			privateSubnet := fmt.Sprintf("grafto-%s-subnet-%v-%v", ctx.Stack(), "private", i+1)

			var pubCidrBlock string
			var priCidrBlock string
			if i == 0 {
				pubCidrBlock = startingSubnetCidrRange
				priCidrBlock = fmt.Sprintf("10.0.%v.0/20", 16)
			} else {
				pubCidrBlock = fmt.Sprintf("10.0.%v.0/20", 16*(i+1))
				priCidrBlock = fmt.Sprintf("10.0.%v.0/20", 16*(i+2))
			}

			public, err := ec2.NewSubnet(ctx, publicSubnet, &ec2.SubnetArgs{
				VpcId:            vpc.ID(),
				CidrBlock:        pulumi.String(pubCidrBlock),
				AvailabilityZone: pulumi.String(az),
			})
			if err != nil {
				return err
			}

			private, err := ec2.NewSubnet(ctx, privateSubnet, &ec2.SubnetArgs{
				VpcId:            vpc.ID(),
				CidrBlock:        pulumi.String(priCidrBlock),
				AvailabilityZone: pulumi.String(az),
			})
			if err != nil {
				return err
			}

			subnets["public"] = append(subnets["public"], public)
			subnets["private"] = append(subnets["private"], private)
		}

		internetGateway, err := ec2.NewInternetGateway(ctx, "grafto-igw", &ec2.InternetGatewayArgs{
			VpcId: vpc.ID(),
		})
		if err != nil {
			return err
		}

		publicRouteTable, err := ec2.NewRouteTable(ctx, "public-route-table", &ec2.RouteTableArgs{
			VpcId: vpc.ID(),
		})
		if err != nil {
			return err
		}

		_, err = ec2.NewRoute(ctx, "public-route", &ec2.RouteArgs{
			DestinationCidrBlock: pulumi.String("0.0.0.0/0"),
			GatewayId:            internetGateway.ID(),
			RouteTableId:         publicRouteTable.ID(),
		})
		if err != nil {
			return err
		}

		publicSubnets := subnets["public"]
		for i, subnet := range publicSubnets {
			_, err = ec2.NewRouteTableAssociation(ctx, fmt.Sprintf("public-route-table-association-%v", i+1), &ec2.RouteTableAssociationArgs{
				RouteTableId: publicRouteTable.ID(),
				SubnetId:     subnet.ID(),
			})
			if err != nil {
				return err
			}
		}

		elasticIP, err := ec2.NewEip(ctx, "grafto-eip", &ec2.EipArgs{})
		if err != nil {
			return err
		}

		natGateway, err := ec2.NewNatGateway(ctx, "grafto-nat", &ec2.NatGatewayArgs{
			AllocationId: elasticIP.AllocationId,
			SubnetId:     subnets["public"][0].ID(),
		})
		if err != nil {
			return err
		}

		privateRouteTable, err := ec2.NewRouteTable(ctx, "private-route-table", &ec2.RouteTableArgs{
			VpcId: vpc.ID(),
		})
		if err != nil {
			return err
		}
		_, err = ec2.NewRoute(ctx, "private-route", &ec2.RouteArgs{
			DestinationCidrBlock: pulumi.String("0.0.0.0/0"),
			NatGatewayId:         natGateway.ID(),
			RouteTableId:         privateRouteTable.ID(),
		})
		if err != nil {
			return err
		}

		privateSubnets := subnets["private"]
		for i, subnet := range privateSubnets {
			_, err = ec2.NewRouteTableAssociation(ctx, fmt.Sprintf("private-route-table-association-%v", i+1), &ec2.RouteTableAssociationArgs{
				RouteTableId: privateRouteTable.ID(),
				SubnetId:     subnet.ID(),
			})
			if err != nil {
				return err
			}
		}

		// Setup security groups
		loadBalancerSg, err := ec2.NewSecurityGroup(ctx, "grafto-load-balancer-sg", &ec2.SecurityGroupArgs{
			VpcId: vpc.ID(),
			Name:  pulumi.String("grafto-load-balancer-sg"),
			Ingress: ec2.SecurityGroupIngressArray{
				&ec2.SecurityGroupIngressArgs{
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
					FromPort:   pulumi.Int(80),
					Protocol:   pulumi.String("tcp"),
					ToPort:     pulumi.Int(80),
				},
			},
			Egress: ec2.SecurityGroupEgressArray{
				&ec2.SecurityGroupEgressArgs{
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
					FromPort:   pulumi.Int(0),
					Protocol:   pulumi.String("-1"),
					ToPort:     pulumi.Int(0),
				},
			},
		})
		if err != nil {
			return err
		}

		// allow all traffic from alb on ephemeral ports
		ecsSg, err := ec2.NewSecurityGroup(ctx, "grafto-ecs-sg", &ec2.SecurityGroupArgs{
			VpcId: vpc.ID(),
			Name:  pulumi.String("grafto-ecs-sg"),
			Ingress: ec2.SecurityGroupIngressArray{
				&ec2.SecurityGroupIngressArgs{
					CidrBlocks:     pulumi.StringArray{pulumi.String("0.0.0.0/0")},
					FromPort:       pulumi.Int(1024),
					Protocol:       pulumi.String("tcp"),
					ToPort:         pulumi.Int(65535),
					SecurityGroups: pulumi.StringArray{loadBalancerSg.ID()},
				},
			},
			Egress: ec2.SecurityGroupEgressArray{
				&ec2.SecurityGroupEgressArgs{
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
					FromPort:   pulumi.Int(0),
					Protocol:   pulumi.String("-1"),
					ToPort:     pulumi.Int(0),
				},
			},
		})
		if err != nil {
			return err
		}

		// Setup load balancing

		applicationLoadBalancer, err := lb.NewLoadBalancer(ctx, "grafto-alb", &lb.LoadBalancerArgs{
			Internal:         pulumi.Bool(false),
			LoadBalancerType: pulumi.String("application"),
			SecurityGroups:   pulumi.StringArray{loadBalancerSg.ID()},
			Subnets:          pulumi.StringArray{subnets["public"][0].ID(), subnets["public"][1].ID()},
		})
		if err != nil {
			return err
		}

		albTargtGroup, err := lb.NewTargetGroup(ctx, "grafto-alb-target-group", &lb.TargetGroupArgs{
			HealthCheck: &lb.TargetGroupHealthCheckArgs{
				Path:     pulumi.String("/health"),
				Protocol: pulumi.String("HTTP"),
			},
			Name:       pulumi.String("grafto-alb-target-group"),
			Port:       pulumi.Int(80),
			Protocol:   pulumi.String("HTTP"),
			TargetType: pulumi.String("ip"),
			VpcId:      vpc.ID(),
		})
		if err != nil {
			return err
		}

		_, err = lb.NewListener(ctx, "grafto-alb-listener", &lb.ListenerArgs{
			LoadBalancerArn: applicationLoadBalancer.Arn,
			Port:            pulumi.Int(80),
			DefaultActions: lb.ListenerDefaultActionArray{
				&lb.ListenerDefaultActionArgs{
					Type:           pulumi.String("forward"),
					TargetGroupArn: albTargtGroup.Arn,
				},
			},
			Protocol: pulumi.String("HTTP"),
		})
		if err != nil {
			return err
		}

		ctx.Export("url", pulumi.Sprintf("http://%s", applicationLoadBalancer.DnsName))

		// Setup ecs
		cluster, err := ecs.NewCluster(ctx, "grafto-cluster", &ecs.ClusterArgs{
			Name: pulumi.String("grafto-cluster"),
		})
		if err != nil {
			return err
		}

		containerDefintion := []containerDefinition{
			{
				Name:  "grafto",
				Image: os.Getenv("GRAFTO_IMAGE"),
				PortMappings: []portMapping{
					{
						ContainerPort: 8000,
						HostPort:      8000,
						Protocol:      "tcp",
					},
				},
				Essential: true,
				Environment: []KeyValuePair{
					{
						Name:  "ENVIRONMENT",
						Value: "production",
					},
					{
						Name:  "SERVER_HOST",
						Value: "0.0.0.0",
					},
					{
						Name:  "SERVER_PORT",
						Value: "8000",
					},
					{
						Name:  "POSTMARK_API_TOKEN",
						Value: "secret-token",
					},
					{
						Name:  "DB_KIND",
						Value: "postgres",
					},
					{
						Name:  "DB_PORT",
						Value: "5432",
					},
					{
						Name:  "DB_HOST",
						Value: os.Getenv("DB_HOST"),
					},
					{
						Name:  "DB_NAME",
						Value: os.Getenv("DB_NAME"),
					},
					{
						Name:  "DB_USER",
						Value: os.Getenv("DB_USER"),
					},
					{
						Name:  "DB_PASSWORD",
						Value: os.Getenv("DB_PASSWORD"),
					},
					{
						Name:  "PASSWORD_PEPPER",
						Value: "secret-pepper",
					},
					{
						Name:  "PROJECT_NAME",
						Value: "Deploy Grafto with Pulumi",
					},
					{
						Name:  "HOST",
						Value: "0.0.0.0:8000",
					},
					{
						Name:  "SCHEME",
						Value: "http",
					},
					{
						Name:  "CSRF_TOKEN",
						Value: "secret-csrf-token",
					},
					{
						Name:  "SESSION_KEY",
						Value: "secret-session-key",
					},
					{
						Name:  "SESSION_ENCRYPTION_KEY",
						Value: "secret-session-encryption-key",
					},
					{
						Name:  "TOKEN_SIGNING_KEY",
						Value: "secret-session-encryption-key",
					},
				},
			},
		}

		jsonData, err := json.Marshal(containerDefintion)
		if err != nil {
			return err
		}

		iamRoleData := iamRolePayload{
			name: "ecs-service-execution-role",
			rolePolicy: policy{
				Version: "2012-10-17",
				Statements: []Statement{
					{
						Action:    []string{"sts:AssumeRole"},
						Principal: map[string]string{"Service": "ecs-tasks.amazonaws.com"},
						Effect:    "Allow",
						Sid:       "EcsServiceRole",
					},
				},
			},
		}

		iamRoleDataJson, err := json.Marshal(iamRoleData.rolePolicy)
		if err != nil {
			return err
		}

		serviceExecutionRole, err := iam.NewRole(ctx, "ecs-role", &iam.RoleArgs{
			AssumeRolePolicy: pulumi.String(string(iamRoleDataJson)),
		})
		if err != nil {
			return err
		}

		serviceExecutionRoleData := policy{
			Version: "2012-10-17",
			Statements: []Statement{
				{
					Action: []string{
						"ecr:*",
					},
					Effect:   "Allow",
					Sid:      "EcsServiceRolePolicy",
					Resource: []string{"*"},
				},
			},
		}
		serviceExecutionRoleDataJson, err := json.Marshal(serviceExecutionRoleData)
		if err != nil {
			return err
		}

		_, err = iam.NewRolePolicy(ctx, "ecs-role-policy", &iam.RolePolicyArgs{
			Name:   pulumi.String("ecs-role-policy"),
			Role:   serviceExecutionRole,
			Policy: pulumi.String(string(serviceExecutionRoleDataJson)),
		})
		if err != nil {
			return err
		}

		taskDefinition, err := ecs.NewTaskDefinition(ctx, "grafto-task-definition", &ecs.TaskDefinitionArgs{
			ContainerDefinitions: pulumi.String(string(jsonData)),
			Cpu:                  pulumi.String("256"),
			ExecutionRoleArn:     serviceExecutionRole.Arn,
			Family:               pulumi.String("grafto"),
			Memory:               pulumi.String("512"),
			NetworkMode:          pulumi.String("awsvpc"),
		})

		_, err = ecs.NewService(ctx, "grafto-service", &ecs.ServiceArgs{
			Cluster:                         cluster.ID(),
			DeploymentMaximumPercent:        pulumi.Int(200),
			DeploymentMinimumHealthyPercent: pulumi.Int(50),
			DesiredCount:                    pulumi.Int(1),
			ForceNewDeployment:              pulumi.Bool(true),
			LoadBalancers: ecs.ServiceLoadBalancerArray{
				&ecs.ServiceLoadBalancerArgs{
					TargetGroupArn: albTargtGroup.Arn,
					ContainerName:  pulumi.String("grafto"),
					ContainerPort:  pulumi.Int(8000),
				},
			},
			NetworkConfiguration: ecs.ServiceNetworkConfigurationArgs{
				AssignPublicIp: pulumi.Bool(true),
				Subnets:        pulumi.StringArray{subnets["private"][0].ID(), subnets["private"][1].ID()},
				SecurityGroups: pulumi.StringArray{ecsSg.ID()},
			},
			Name:           pulumi.String("grafto-service"),
			LaunchType:     pulumi.String("FARGATE"),
			TaskDefinition: taskDefinition.Arn,
		})
		if err != nil {
			return err
		}

		return nil
	})
}

type portMapping struct {
	ContainerPort int32  `json:"containerPort"`
	HostPort      int32  `json:"hostPort"`
	Protocol      string `json:"protocol"`
}
type KeyValuePair struct {
	Value string `json:"value"`
	Name  string `json:"name"`
}

type healthCheck struct {
	Command     []string `json:"command"`
	Interval    int      `json:"interval"`
	Retries     int      `json:"retries"`
	StartPeriod int      `json:"startPeriod"`
	TimeOut     int      `json:"timeOut"`
}

type containerDefinition struct {
	Name         string         `json:"name"`
	Image        string         `json:"image"`
	PortMappings []portMapping  `json:"portMappings"`
	Essential    bool           `json:"essential"`
	Environment  []KeyValuePair `json:"environment"`
}

type Statement struct {
	Action    []string          `json:"Action"`
	Principal map[string]string `json:"Principal,omitempty"`
	Effect    string            `json:"Effect"`
	Sid       string            `json:"Sid"`
	Resource  []string          `json:"Resource,omitempty"`
}

type policy struct {
	Version    string      `json:"Version"`
	Statements []Statement `json:"Statement"`
}

type iamRolePayload struct {
	name       string
	rolePolicy policy
}

type iamRolePolicyPayload struct {
	name        string
	iamRoleName pulumi.StringOutput
	rolePolicy  policy
}
