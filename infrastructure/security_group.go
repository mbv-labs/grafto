package infrastructure

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type SecurityGroupSetting struct {
	CidrBlocks []string
	FromPort   int32
	ToPort     int32
	Protocol   string
}

func CreateSecurityGroup(
	ctx *pulumi.Context,
	name string,
	vpcID pulumi.IDOutput,
	ingressRule SecurityGroupSetting,
	egressRule SecurityGroupSetting,
) (*ec2.SecurityGroup, error) {
	securityGroup, err := ec2.NewSecurityGroup(ctx, name, &ec2.SecurityGroupArgs{
		VpcId: vpcID,
		Name:  pulumi.String(name),
		Ingress: ec2.SecurityGroupIngressArray{
			ec2.SecurityGroupIngressArgs{
				CidrBlocks: pulumi.ToStringArray(ingressRule.CidrBlocks),
				FromPort:   pulumi.Int(ingressRule.FromPort),
				Protocol:   pulumi.String(ingressRule.Protocol),
				ToPort:     pulumi.Int(ingressRule.ToPort),
			},
		},
		Egress: ec2.SecurityGroupEgressArray{
			ec2.SecurityGroupEgressArgs{
				CidrBlocks: pulumi.ToStringArray(egressRule.CidrBlocks),
				FromPort:   pulumi.Int(egressRule.FromPort),
				Protocol:   pulumi.String(egressRule.Protocol),
				ToPort:     pulumi.Int(egressRule.ToPort),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return securityGroup, nil
}
