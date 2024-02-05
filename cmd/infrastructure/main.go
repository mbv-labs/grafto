package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var availabilityZones = []string{"eu-central-1a", "eu-central-1b"}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create an AWS resource (S3 Bucket)

		// Setup networking
		vpc, err := ec2.NewVpc(ctx, "grafto-vpc", &ec2.VpcArgs{
			CidrBlock:          pulumi.String("0.0.0.0/16"),
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

		privateSubnets := subnets["private"]
		for i, subnet := range privateSubnets {
			_, err := ec2.NewRoute(ctx, fmt.Sprintf("private-route-%v", i+1), &ec2.RouteArgs{
				DestinationCidrBlock: pulumi.String("0.0.0.0/16"),
				NatGatewayId:         natGateway.ID(),
				RouteTableId:         privateRouteTable.ID(),
			})
			if err != nil {
				return err
			}

			_, err = ec2.NewRouteTableAssociation(ctx, fmt.Sprintf("private-route-table-association-%v", i+1), &ec2.RouteTableAssociationArgs{
				RouteTableId: privateRouteTable.ID(),
				SubnetId:     subnet.ID(),
			})
			if err != nil {
				return err
			}
		}

		// Setup load balancing

		// Setup ecs

		return nil
	})
}
