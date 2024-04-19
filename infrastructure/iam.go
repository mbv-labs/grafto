package infrastructure

import (
	"encoding/json"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateServiceLinkedRole(ctx *pulumi.Context, name, desc, service string) error {
	_, err := iam.NewServiceLinkedRole(ctx, name, &iam.ServiceLinkedRoleArgs{
		AwsServiceName: pulumi.String(service),
		Description:    pulumi.String(desc),
	})
	if err != nil {
		return err
	}

	return nil
}

type RoleStatement struct {
	Action    []string            `json:"Action"`
	Principal map[string][]string `json:"Principal"`
	Effect    string              `json:"Effect"`
	Sid       string              `json:"Sid"`
}

type PolicyStatement struct {
	Action   []string `json:"Action"`
	Effect   string   `json:"Effect"`
	Sid      string   `json:"Sid"`
	Resource []string `json:"Resource"`
}

func CreateIamRoleWithPolicy(
	ctx *pulumi.Context,
	roleName string,
	policyName string,
	roleStatements []RoleStatement,
	policyStatements []PolicyStatement,
) (*iam.Role, error) {
	rolePayload := map[string]any{
		"Version":   "2012-10-17",
		"Statement": roleStatements,
		"Condition": "Allow",
	}

	jsonifiedRole, err := json.Marshal(rolePayload)
	if err != nil {
		return nil, err
	}

	role, err := iam.NewRole(ctx, roleName, &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(jsonifiedRole),
		Name:             pulumi.String(roleName),
	})
	if err != nil {
		return nil, err
	}

	policyPayload := map[string]any{
		"Version":   "2012-10-17",
		"Statement": policyStatements,
		"Condition": "Allow",
	}

	jsonifiedPolicy, err := json.Marshal(policyPayload)
	if err != nil {
		return nil, err
	}

	_, err = iam.NewRolePolicy(ctx, policyName, &iam.RolePolicyArgs{
		Name:   pulumi.String(policyName),
		Policy: pulumi.String(jsonifiedPolicy),
		Role:   role,
	})
	if err != nil {
		return nil, err
	}

	return role, nil
}
