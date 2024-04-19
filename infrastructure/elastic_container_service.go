package infrastructure

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateEcsCluster(ctx *pulumi.Context, name string) (*ecs.Cluster, error) {
	cluster, err := ecs.NewCluster(ctx, name, &ecs.ClusterArgs{
		Name: pulumi.String(name),
	})
	if err != nil {
		return nil, err
	}

	return cluster, nil
}

func CreateEcsClusterService(
	ctx *pulumi.Context,
	name string,
	cluster *ecs.Cluster,
	securityGroupsIDs []pulumi.IDOutput,
	subnetsIDs []pulumi.IDOutput,
	taskDefinitionArn pulumi.StringOutput,
	taskName string,
	taskPort int,
	targetGroupArn pulumi.StringOutput,
) (*ecs.Service, error) {
	securityGroups := make(pulumi.StringArray, len(securityGroupsIDs))
	for i, securityGroup := range securityGroups {
		securityGroups[i] = securityGroup
	}

	subnets := make(pulumi.StringArray, len(subnetsIDs))
	for i, subnet := range subnets {
		subnets[i] = subnet
	}

	service, err := ecs.NewService(ctx, name, &ecs.ServiceArgs{
		Cluster:      cluster.Arn,
		DesiredCount: pulumi.Int(2),
		LaunchType:   pulumi.String("FARGATE"),
		Name:         pulumi.String(name),
		NetworkConfiguration: &ecs.ServiceNetworkConfigurationArgs{
			AssignPublicIp: pulumi.Bool(true),
			SecurityGroups: securityGroups,
			Subnets:        subnets,
		},
		PlatformVersion: pulumi.String("1.4.0"),
		TaskDefinition:  taskDefinitionArn,
		LoadBalancers: ecs.ServiceLoadBalancerArray{
			ecs.ServiceLoadBalancerArgs{
				ContainerName:  pulumi.String(taskName),
				ContainerPort:  pulumi.Int(taskPort),
				TargetGroupArn: targetGroupArn,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return service, nil
}

type ContainerDefinition struct {
	Name        string
	ImageUri    string
	PortMapping string
	Environment string
}

func CreateTaskDefinition(ctx *pulumi.Context, taskName string) {
	task, err := ecs.NewTaskDefinition(ctx, taskName, &ecs.TaskDefinitionArgs{
		ContainerDefinitions:    nil,
		Cpu:                     nil,
		EphemeralStorage:        nil,
		ExecutionRoleArn:        nil,
		Family:                  nil,
		InferenceAccelerators:   nil,
		IpcMode:                 nil,
		Memory:                  nil,
		NetworkMode:             nil,
		PidMode:                 nil,
		PlacementConstraints:    nil,
		ProxyConfiguration:      nil,
		RequiresCompatibilities: nil,
		RuntimePlatform:         nil,
		SkipDestroy:             nil,
		Tags:                    nil,
		TaskRoleArn:             nil,
		TrackLatest:             nil,
		Volumes:                 nil,
	})
}
