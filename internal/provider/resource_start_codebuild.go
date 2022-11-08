package toluna

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceStartCodebuild() *schema.Resource {
	return &schema.Resource{
		Create: resourceStartCodebuildCreate,
		Read:   resourceStartCodebuildRead,
		Update: resourceStartCodebuildUpdate,
		Delete: resourceStartCodebuildDelete,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"aws_profile": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"environment_variables": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},

						"value": {
							Type:     schema.TypeString,
							Required: true,
						},

						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
								var result bool = false
								v := val.(string)
								arr := [3]string{"PLAINTEXT", "SECRETS_MANAGER", "PARAMETER_STORE"}
								for _, x := range arr {
									if x == v {
										result = true
										break
									}
								}
								if !result {
									errs = append(errs, fmt.Errorf("%q must Equal one of: [PLAINTEXT | SECRETS_MANAGER | PARAMETER_STORE] , got: %s", key, v))
								}
								return
							},
						},
					},
				},
			},
		},
	}
}

func startCodebuild(d *schema.ResourceData, m interface{}, action string) (str string, er error) {
	var profile = d.Get("aws_profile").(string)
	if profile == "" {
		profile = "default"
	}
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           profile,
	}))

	client := codebuild.New(sess, &aws.Config{Region: aws.String(d.Get("region").(string))})
	envVar := []*codebuild.EnvironmentVariable{}
	param := &codebuild.EnvironmentVariable{
		Name:  aws.String("action"),
		Type:  aws.String("PLAINTEXT"),
		Value: aws.String(action)}
	envVar = append(envVar, param)
	envVars := d.Get("environment_variables").(*schema.Set)
	envVars_list := envVars.List()
	for i := range envVars_list {
		envVars_map := envVars_list[i].(map[string]interface{})
		param := &codebuild.EnvironmentVariable{
			Name:  aws.String(envVars_map["name"].(string)),
			Type:  aws.String(envVars_map["type"].(string)),
			Value: aws.String(envVars_map["value"].(string))}
		envVar = append(envVar, param)
	}
	input := &codebuild.StartBuildInput{
		ProjectName:                  aws.String(d.Get("project_name").(string)),
		EnvironmentVariablesOverride: envVar,
	}
	result, err := client.StartBuild(input)
	if err != nil {
		return "", fmt.Errorf("Error calling build project:%s", err)
	}
	Ids := []*string{}
	Ids = append(Ids, result.Build.Id)
	for {
		buildstatus, err := client.BatchGetBuilds(&codebuild.BatchGetBuildsInput{Ids: Ids})
		if err != nil {
			break
		}
		if *buildstatus.Builds[0].BuildComplete {
			break
		}
	}

	buildresult, err := client.BatchGetBuilds(&codebuild.BatchGetBuildsInput{Ids: Ids})
	if err != nil {
		return "", fmt.Errorf("Error failed getting build:%s", *buildresult.Builds[0].Id)
	}

	if *buildresult.Builds[0].BuildStatus != "SUCCEEDED" {
		return "", fmt.Errorf("Error build status:%s,%s", *buildresult.Builds[0].BuildStatus, *buildresult.Builds[0].Id)
	}
	return *buildresult.Builds[0].Arn, nil
}

func resourceStartCodebuildCreate(d *schema.ResourceData, m interface{}) error {
	result, err := startCodebuild(d, m, "apply")
	if err != nil {
		d.Partial(true)
		return err
	}
	d.SetId(result)
	return resourceStartCodebuildRead(d, m)
}

func resourceStartCodebuildRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceStartCodebuildUpdate(d *schema.ResourceData, m interface{}) error {
	result, err := startCodebuild(d, m, "apply")
	if err != nil {
		d.Partial(true)
		return err
	}
	d.SetId(result)
	return resourceStartCodebuildRead(d, m)
}

func resourceStartCodebuildDelete(d *schema.ResourceData, m interface{}) error {
	result, err := startCodebuild(d, m, "destroy")
	if err != nil {
		d.Partial(true)
		return err
	}
	d.SetId(result)
	return resourceStartCodebuildRead(d, m)
}
