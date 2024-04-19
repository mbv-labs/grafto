package main

import (
	"github.com/mbv-labs/grafto/infrastructure"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		network, err := infrastructure.CreateNetwork(ctx)
		if err != nil {
			return err
		}

		securityGroup, err := infrastructure.CreateSecurityGroup(
			ctx,
			"grafto-security-group",
			network.VpcInstance.ID(),
			infrastructure.SecurityGroupSetting{
				CidrBlocks: []string{
					"0.0.0.0/0",
				},
				FromPort: 80,
				ToPort:   8080,
				Protocol: "tcp",
			},
			infrastructure.SecurityGroupSetting{
				CidrBlocks: []string{
					"0.0.0.0/0",
				},
				FromPort: 0,
				ToPort:   0,
				Protocol: "-1",
			},
		)
		if err != nil {
			return err
		}

		if err := infrastructure.CreateServiceLinkedRole(
			ctx,
			"elastic-container-service",
			"allows access to ecs",
			"ecs.amazon.com",
		); err != nil {
			return err
		}

		if err := infrastructure.CreateServiceLinkedRole(
			ctx,
			"rds",
			"allows access to rds",
			"rds.amazon.com",
		); err != nil {
			return err
		}

		roleWithPolicy, err := infrastructure.CreateIamRoleWithPolicy(
			ctx,
			"grafto-role",
			"grafto-role-policy",
			[]infrastructure.RoleStatement{
				{
					Action: []string{
						"sts:AssumeRole",
					},
					Principal: map[string][]string{
						"Service": {
							"ecs-tasks.amazonaws.com",
							"ecs.amazonaws.com",
						},
					},
					Effect: "Allow",
					Sid:    "GraftoRole",
				},
			},
			[]infrastructure.PolicyStatement{
				{
					Action: []string{
						"ecs:*",
						"ssm:*",
						"rds:*",
					},
					Effect:   "Allow",
					Sid:      "GraftoRolePolicy",
					Resource: []string{"*"},
				},
			},
		)

		return nil
	})
}
