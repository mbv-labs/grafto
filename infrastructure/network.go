package infrastructure

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	availabilityZones = []string{"eu-central-1a", "eu-central-1b"}

	vpcCidrBlock = "10.0.0.0/16"

	publicSubnetOne  = "10.0.0.0/18"
	privateSubnetOne = "10.0.64.0/18"
	publicSubnetTwo  = "10.0.128.0/18"
	privateSubnetTwo = "10.0.192.0/18"
)

type Network struct {
	VpcInstance *ec2.Vpc
	Subnets     map[string][]*ec2.Subnet
}

func CreateNetwork(ctx *pulumi.Context) (Network, error) {
	vpc, err := ec2.NewVpc(ctx, "vpc", &ec2.VpcArgs{
		CidrBlock:          pulumi.String(vpcCidrBlock),
		EnableDnsHostnames: pulumi.Bool(true),
		EnableDnsSupport:   pulumi.Bool(true),
	})
	if err != nil {
		return Network{}, err
	}

	subnets := make(map[string][]*ec2.Subnet, 2)
	for i, az := range availabilityZones {
		privateCidrBlock := ""
		if i%2 == 0 {
			privateCidrBlock = privateSubnetOne
		}
		if i%2 == 1 {
			privateCidrBlock = privateSubnetTwo
		}

		privateSubnet, err := ec2.NewSubnet(
			ctx,
			fmt.Sprintf("private-subnet-%v", i+1),
			&ec2.SubnetArgs{
				AvailabilityZone: pulumi.String(az),
				CidrBlock:        pulumi.String(privateCidrBlock),
				VpcId:            vpc.ID(),
			},
		)
		if err != nil {
			return Network{}, err
		}

		subnets["private"] = append(subnets["private"], privateSubnet)

		publicCidrBlock := ""
		if i%2 == 0 {
			publicCidrBlock = publicSubnetOne
		}
		if i%2 == 1 {
			publicCidrBlock = publicSubnetTwo
		}

		publicSubnet, err := ec2.NewSubnet(
			ctx,
			fmt.Sprintf("public-subnet-%v", i+1),
			&ec2.SubnetArgs{
				AvailabilityZone: pulumi.String(az),
				CidrBlock:        pulumi.String(publicCidrBlock),
				VpcId:            vpc.ID(),
			},
		)
		if err != nil {
			return Network{}, err
		}

		subnets["public"] = append(subnets["public"], publicSubnet)
	}

	if err := SetupInternetGateway(ctx, vpc.ID(), subnets["public"]); err != nil {
		return Network{}, err
	}

	if err := SetupNatGateway(ctx, vpc.ID(), subnets["public"][0].ID(), subnets["private"]); err != nil {
		return Network{}, err
	}

	return Network{
		VpcInstance: vpc,
	}, nil
}

func SetupInternetGateway(
	ctx *pulumi.Context,
	vpcID pulumi.IDOutput,
	publicSubnets []*ec2.Subnet,
) error {
	igw, err := ec2.NewInternetGateway(ctx, "internet-gateway", &ec2.InternetGatewayArgs{
		VpcId: vpcID,
	})
	if err != nil {
		return err
	}

	routeTable, err := ec2.NewRouteTable(ctx, "igw-route-table", &ec2.RouteTableArgs{
		VpcId: vpcID,
	})
	if err != nil {
		return err
	}

	_, err = ec2.NewRoute(ctx, "igw-route", &ec2.RouteArgs{
		DestinationCidrBlock: pulumi.String("0.0.0.0/0"),
		GatewayId:            igw.ID(),
		RouteTableId:         routeTable.ID(),
	})
	if err != nil {
		return err
	}

	for i, subnet := range publicSubnets {
		_, err = ec2.NewRouteTableAssociation(
			ctx,
			fmt.Sprintf(
				"public-table-association-%v",
				i+1,
			),
			&ec2.RouteTableAssociationArgs{
				RouteTableId: igw.ID(),
				SubnetId:     subnet.ID(),
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func SetupNatGateway(
	ctx *pulumi.Context,
	vpcID pulumi.IDOutput,
	publicSubnetID pulumi.IDOutput,
	privateSubnets []*ec2.Subnet,
) error {
	elasticIP, err := ec2.NewEip(
		ctx,
		"elasticIP",
		&ec2.EipArgs{},
	)
	if err != nil {
		return err
	}

	ngw, err := ec2.NewNatGateway(ctx, "nat-gateway", &ec2.NatGatewayArgs{
		SubnetId:     publicSubnetID,
		AllocationId: elasticIP.ID(),
	})
	if err != nil {
		return err
	}

	routeTable, err := ec2.NewRouteTable(ctx, "ngw-route-table", &ec2.RouteTableArgs{
		VpcId: vpcID,
	})
	if err != nil {
		return err
	}

	_, err = ec2.NewRoute(ctx, "ngw-route", &ec2.RouteArgs{
		DestinationCidrBlock: pulumi.String("0.0.0.0/0"),
		NatGatewayId:         ngw.ID(),
		RouteTableId:         routeTable.ID(),
	})
	if err != nil {
		return err
	}

	for i, subnet := range privateSubnets {
		_, err = ec2.NewRouteTableAssociation(
			ctx,
			fmt.Sprintf(
				"private-table-association-%v",
				i+1,
			),
			&ec2.RouteTableAssociationArgs{
				RouteTableId: ngw.ID(),
				SubnetId:     subnet.ID(),
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}
