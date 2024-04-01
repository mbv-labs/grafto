package awsecspulumi

import (
	"errors"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
)

var availabilityZones = []string{"eu-central-1a", "eu-central-1b"}

type NetworkBuilder struct {
	Network *Network
}

type Network struct {
	Name            string
	VPC             VPC
	PublicSubnets   []Subnet
	PrivateSubnets  []Subnet
	InternetGateway InternetGateway
	NatGateway      InternetGateway
}

func NewNetworkBuilder() NetworkBuilder {
	return NetworkBuilder{}
}

func (nb NetworkBuilder) Defaults() NetworkBuilder {
	return nb
}

func (nb NetworkBuilder) AddVPC(vpc VPC) NetworkBuilder {
	nb.Network.VPC = vpc
	return nb
}

func (nb NetworkBuilder) AddSubnet(subnet Subnet) NetworkBuilder {
	if subnet.public {
		for _, s := range nb.Network.PublicSubnets {
			if s.cidrBlock == subnet.cidrBlock {
				panic(errors.New("cidr range already in use"))
			}
		}

		nb.Network.PublicSubnets = append(nb.Network.PublicSubnets, subnet)
	}

	if !subnet.public {
		for _, s := range nb.Network.PrivateSubnets {
			if s.cidrBlock == subnet.cidrBlock {
				panic(errors.New("cidr range already in use"))
			}
		}

		nb.Network.PrivateSubnets = append(nb.Network.PrivateSubnets, subnet)
	}

	return nb
}

func (nb NetworkBuilder) AddInternetGateway(gateway InternetGateway) NetworkBuilder {
	return nb
}

func (nb NetworkBuilder) AddInternetGateway(gateway InternetGateway) NetworkBuilder {}

func (nb NetworkBuilder) Build(ctx *pulumi.Context) *Network {
	return nb.Network
}

type (
	VPC struct {
		instance  *ec2.VPC
		cidrBlock string
	}
	Subnet struct {
		instance         *ec2.Subnet
		availabilityZone string
		cidrBlock        string
		public           bool
	}
	InternetGateway struct {
		instance *ec2.InternetGateway
		name     string
	}
	NatGateway struct {
		instance *ec2.InternetGateway
		name     string
	}
)
