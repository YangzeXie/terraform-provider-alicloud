package alicloud

import (
	"fmt"
	"github.com/PaesslerAG/jsonpath"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAlicloudMseCluster() *schema.Resource {
	return &schema.Resource{
		Create:        resourceAlicloudMseClusterCreate,
		Read:          resourceAlicloudMseClusterRead,
		Update:        resourceAlicloudMseClusterUpdate,
		Delete:        resourceAlicloudMseClusterDelete,
		CustomizeDiff: customizeMseClusterDiff,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"acl_entry_list": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if d.Get("pub_network_flow").(string) == "0" && d.Get("net_type").(string) == "privatenet" {
						return true
					}
					return false
				},
			},
			"cluster_alias_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cluster_specification": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: StringInSlice([]string{"MSE_SC_1_2_60_c", "MSE_SC_2_4_60_c", "MSE_SC_4_8_60_c", "MSE_SC_8_16_60_c", "MSE_SC_16_32_60_c", "MSE_SC_1_2_200_c", "MSE_SC_2_4_200_c", "MSE_SC_4_8_200_c", "MSE_SC_8_16_200_c", "MSE_SC_16_32_200_c", "MSE_SC_SERVERLESS"}, false),
			},
			"cluster_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: StringInSlice([]string{"Eureka", "Nacos-Ans", "ZooKeeper"}, false),
			},
			"cluster_version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"version_code": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				// 具体版本号，可以是 LATEST 或具体版本号
			},
			"connection_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: StringInSlice([]string{"slb", "single_eni"}, false),
			},
			"disk_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"instance_count": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"net_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: StringInSlice([]string{"privatenet", "pubnet", "both"}, false),
			},
			"payment_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: StringInSlice([]string{"PayAsYouGo", "Subscription"}, false),
			},
			"tags": tagsSchema(),
			"resource_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"private_slb_specification": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"pub_network_flow": {
				Type:     schema.TypeString,
				Required: true,
			},
			"pub_slb_specification": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"request_pars": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"vswitch_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"mse_version": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: StringInSlice([]string{"mse_dev", "mse_basic", "mse_pro", "mse_serverless"}, false),
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"app_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAlicloudMseClusterCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	mseService := MseService{client}
	var response map[string]interface{}
	action := "CreateCluster"
	request := make(map[string]interface{})
	var err error
	request["ClusterSpecification"] = d.Get("cluster_specification")
	request["ClusterType"] = d.Get("cluster_type")
	request["ClusterVersion"] = d.Get("cluster_version")
	if v, ok := d.GetOk("connection_type"); ok {
		request["ConnectionType"] = v
	}
	if v, ok := d.GetOk("disk_type"); ok {
		request["DiskType"] = v
	}

	request["InstanceCount"] = d.Get("instance_count")
	request["NetType"] = d.Get("net_type")
	if v, ok := d.GetOk("private_slb_specification"); ok {
		request["PrivateSlbSpecification"] = v
	}

	request["PubNetworkFlow"] = d.Get("pub_network_flow")

	if v, ok := d.GetOk("pub_slb_specification"); ok {
		request["PubSlbSpecification"] = v
	}
	if v, ok := d.GetOk("mse_version"); ok {
		request["MseVersion"] = v
	}
	if v, ok := d.GetOk("request_pars"); ok {
		request["RequestPars"] = v
	}

	if v, ok := d.GetOk("vpc_id"); ok {
		request["VpcId"] = v
	}

	if v, ok := d.GetOk("vswitch_id"); ok {
		request["VSwitchId"] = v
	}
	if v, ok := d.GetOk("payment_type"); ok {
		request["ChargeType"] = convertMsePaymentTypeToChargeType(v)
	}

	request["Region"] = client.RegionId

	if request["VpcId"] == nil && request["VSwitchId"] != nil {
		vpcService := VpcService{client}
		vsw, err := vpcService.DescribeVswitch(request["VSwitchId"].(string))
		if err != nil {
			return WrapError(err)
		}
		if v, ok := request["VpcId"].(string); !ok || v == "" {
			request["VpcId"] = vsw["VpcId"]
		}
	}
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(client.GetRetryTimeout(d.Timeout(schema.TimeoutCreate)), func() *resource.RetryError {
		response, err = client.RpcPost("mse", "2019-05-31", action, nil, request, false)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, response, request)
		return nil
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_mse_cluster", action, AlibabaCloudSdkGoERROR)
	}
	if fmt.Sprint(response["Success"]) == "false" {
		return WrapError(fmt.Errorf("%s failed, response: %v", action, response))
	}
	d.SetId(fmt.Sprint(response["InstanceId"]))
	stateConf := BuildStateConf([]string{}, []string{"INIT_SUCCESS"}, d.Timeout(schema.TimeoutCreate), 60*time.Second, mseService.MseClusterStateRefreshFunc(d.Id(), []string{"INIT_FAILED"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}
	return resourceAlicloudMseClusterUpdate(d, meta)
}

func resourceAlicloudMseClusterRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	mseService := MseService{client}
	object, err := mseService.DescribeMseCluster(d.Id())
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_mse_cluster mseService.DescribeMseCluster Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("cluster_type", object["ClusterType"])
	d.Set("cluster_specification", object["ClusterSpecification"])
	d.Set("instance_count", formatInt(object["InstanceCount"]))
	d.Set("pub_network_flow", object["PubNetworkFlow"])
	d.Set("mse_version", object["MseVersion"])
	d.Set("net_type", object["NetType"])
	d.Set("vswitch_id", object["VSwitchId"])
	d.Set("cluster_version", object["OrderClusterVersion"])
	d.Set("version_code", object["ClusterVersion"])
	d.Set("cluster_alias_name", object["ClusterAliasName"])
	d.Set("connection_type", object["ConnectionType"])
	d.Set("vpc_id", object["VpcId"])
	d.Set("cluster_id", object["ClusterId"])
	d.Set("app_version", object["AppVersion"])
	d.Set("status", object["InitStatus"])
	d.Set("resource_group_id", object["ResourceGroupId"])
	d.Set("payment_type", convertMseChargeTypeToPaymentType(object["ChargeType"]))
	tagsMaps, _ := jsonpath.Get("$.Tags", object)
	d.Set("tags", tagsToMap(tagsMaps))

	return nil
}

func resourceAlicloudMseClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	mseService := MseService{client}
	var response map[string]interface{}
	var err error
	d.Partial(true)

	update := false
	action := "ChangeResourceGroup"
	request := make(map[string]interface{})

	request["ResourceId"] = d.Id()
	request["ResourceRegionId"] = client.RegionId
	if d.HasChange("resource_group_id") {
		update = true
		request["ResourceGroupId"] = d.Get("resource_group_id")
	}

	request["ResourceType"] = "Cluster"
	if update {
		wait := incrementalWait(3*time.Second, 5*time.Second)
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			response, err = client.RpcPost("mse", "2019-05-31", action, nil, request, false)

			if err != nil {
				if NeedRetry(err) {
					wait()
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			addDebug(action, response, request)
			return nil
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}
		d.SetPartial("resource_group_id")
	}

	if d.HasChange("acl_entry_list") {
		request := map[string]interface{}{
			"InstanceId": d.Id(),
		}
		request["AclEntryList"] = convertListToCommaSeparate(d.Get("acl_entry_list").(*schema.Set).List())
		action := "UpdateAcl"
		wait := incrementalWait(3*time.Second, 3*time.Second)
		err = resource.Retry(client.GetRetryTimeout(d.Timeout(schema.TimeoutUpdate)), func() *resource.RetryError {
			response, err = client.RpcPost("mse", "2019-05-31", action, nil, request, false)
			if err != nil {
				if NeedRetry(err) {
					wait()
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			addDebug(action, response, request)
			return nil
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}
		if fmt.Sprint(response["Success"]) == "false" {
			return WrapErrorf(fmt.Errorf("%s failed, response: %v", action, response), DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}
		d.SetPartial("acl_entry_list")
	}
	update = false
	request = map[string]interface{}{
		"InstanceId": d.Id(),
	}
	if d.HasChange("cluster_alias_name") {
		update = true
		request["ClusterAliasName"] = d.Get("cluster_alias_name")
	}
	if update {
		action := "UpdateCluster"
		wait := incrementalWait(3*time.Second, 3*time.Second)
		err = resource.Retry(client.GetRetryTimeout(d.Timeout(schema.TimeoutUpdate)), func() *resource.RetryError {
			response, err = client.RpcPost("mse", "2019-05-31", action, nil, request, false)
			if err != nil {
				if NeedRetry(err) {
					wait()
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			addDebug(action, response, request)
			return nil
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}
		if fmt.Sprint(response["Success"]) == "false" {
			return WrapErrorf(fmt.Errorf("%s failed, response: %v", action, response), DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}
		d.SetPartial("cluster_alias_name")
	}

	update = false
	object, err := mseService.DescribeMseCluster(d.Id())
	updateClusterSpecReq := map[string]interface{}{
		"InstanceId": d.Id(),
		"ClusterId":  object["ClusterId"],
	}

	if !d.IsNewResource() && d.HasChange("cluster_specification") {
		update = true
	}
	if v, ok := d.GetOk("cluster_specification"); ok {
		updateClusterSpecReq["ClusterSpecification"] = v
	}

	if !d.IsNewResource() && d.HasChange("instance_count") {
		update = true
	}
	if v, ok := d.GetOk("instance_count"); ok {
		updateClusterSpecReq["InstanceCount"] = v
	}

	if update {
		action := "UpdateClusterSpec"
		wait := incrementalWait(3*time.Second, 3*time.Second)
		err = resource.Retry(client.GetRetryTimeout(d.Timeout(schema.TimeoutUpdate)), func() *resource.RetryError {
			response, err = client.RpcPost("mse", "2019-05-31", action, nil, updateClusterSpecReq, false)
			if err != nil {
				if NeedRetry(err) {
					wait()
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			addDebug(action, response, updateClusterSpecReq)
			return nil
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}
		if fmt.Sprint(response["Success"]) == "false" {
			return WrapErrorf(fmt.Errorf("%s failed, response: %v", action, response), DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}

		stateConf := BuildStateConf([]string{}, []string{"SCALE_SUCCESS"}, d.Timeout(schema.TimeoutUpdate), 60*time.Second, mseService.MseClusterStateRefreshFunc(d.Id(), []string{"INIT_FAILED"}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}

		d.SetPartial("cluster_specification")
		d.SetPartial("instance_count")
	}

	update = false
	if d.HasChange("tags") {
		update = true
		mseServiceV2 := MseService{client}
		if err := mseServiceV2.SetResourceTags(d, "CLUSTER"); err != nil {
			return WrapError(err)
		}
		d.SetPartial("tags")
	}

	if !d.IsNewResource() && d.HasChanges("vpc_id", "vswitch_id") {
		update = true
		request := map[string]interface{}{
			"InstanceId": d.Id(),
		}
		request["VpcId"] = d.Get("vpc_id")
		request["VswId"] = d.Get("vswitch_id")
		action := "UpdateClusterNetwork"
		wait := incrementalWait(3*time.Second, 3*time.Second)
		err = resource.Retry(client.GetRetryTimeout(d.Timeout(schema.TimeoutUpdate)), func() *resource.RetryError {
			response, err = client.RpcPost("mse", "2019-05-31", action, nil, request, false)
			if err != nil {
				if NeedRetry(err) {
					wait()
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			addDebug(action, response, request)
			return nil
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}
		if fmt.Sprint(response["Success"]) == "false" {
			return WrapErrorf(fmt.Errorf("%s failed, response: %v", action, response), DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}

		stateConf := BuildStateConf([]string{}, []string{"SCALE_SUCCESS"}, d.Timeout(schema.TimeoutUpdate), 60*time.Second, mseService.MseClusterStateRefreshFunc(d.Id(), []string{"INIT_FAILED"}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}

		d.SetPartial("vpc_id")
		d.SetPartial("vswitch_id")
	}

	if !d.IsNewResource() && d.HasChange("pub_network_flow") {
		update = true
		updateNetworkFlowRequest := map[string]interface{}{
			"InstanceId": d.Id(),
		}
		updateNetworkFlowRequest["PubNetworkFlow"] = d.Get("pub_network_flow")
		updateNetworkFlowRequest["AutoPay"] = true

		action := "UpdateClusterSpec"
		wait := incrementalWait(3*time.Second, 3*time.Second)
		err = resource.Retry(client.GetRetryTimeout(d.Timeout(schema.TimeoutUpdate)), func() *resource.RetryError {
			response, err = client.RpcPost("mse", "2019-05-31", action, nil, updateNetworkFlowRequest, false)
			if err != nil {
				if NeedRetry(err) {
					wait()
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			addDebug(action, response, updateNetworkFlowRequest)
			return nil
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}
		if fmt.Sprint(response["Success"]) == "false" {
			return WrapErrorf(fmt.Errorf("%s failed, response: %v", action, response), DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}

		stateConf := BuildStateConf([]string{}, []string{"SCALE_SUCCESS"}, d.Timeout(schema.TimeoutUpdate), 60*time.Second, mseService.MseClusterStateRefreshFunc(d.Id(), []string{"INIT_FAILED"}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}

		d.SetPartial("pub_network_flow")
	}

	if !d.IsNewResource() && d.HasChange("version_code") {
		update = true
		targetVersion := d.Get("version_code").(string)
		currentVersion := object["ClusterVersion"].(string)

		// 如果当前版本不是目标版本，执行更新
		if currentVersion != targetVersion {
			updateRequest := map[string]interface{}{
				"ClusterId":   object["ClusterId"],
				"VersionCode": targetVersion,
			}

			action := "UpdateImage"
			wait := incrementalWait(3*time.Second, 3*time.Second)
			err = resource.Retry(client.GetRetryTimeout(d.Timeout(schema.TimeoutUpdate)), func() *resource.RetryError {
				response, err = client.RpcPost("mse", "2019-05-31", action, nil, updateRequest, false)
				if err != nil {
					if NeedRetry(err) {
						wait()
						return resource.RetryableError(err)
					}
					return resource.NonRetryableError(err)
				}
				addDebug(action, response, updateRequest)
				return nil
			})
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
			}
			if fmt.Sprint(response["Success"]) == "false" {
				return WrapErrorf(fmt.Errorf("%s failed, response: %v", action, response), DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
			}

			stateConf := BuildStateConf([]string{}, []string{"SCALE_SUCCESS"}, d.Timeout(schema.TimeoutUpdate), 60*time.Second, mseService.MseClusterStateRefreshFunc(d.Id(), []string{"INIT_FAILED"}))
			if _, err := stateConf.WaitForState(); err != nil {
				return WrapErrorf(err, IdMsg, d.Id())
			}

			d.SetPartial("version_code")
		} else {
			update = false
		}
	}

	d.Partial(false)
	return resourceAlicloudMseClusterRead(d, meta)
}

func resourceAlicloudMseClusterDelete(d *schema.ResourceData, meta interface{}) error {

	if v, ok := d.GetOk("payment_type"); ok {
		if v == "Subscription" {
			log.Printf("[WARN] Cannot destroy resource alicloud_mse_cluster which payment_type valued Subscription. Terraform will remove this resource from the state file, however resources may remain.")
			return nil
		}
	}

	client := meta.(*connectivity.AliyunClient)
	mseService := MseService{client}
	action := "DeleteCluster"
	var response map[string]interface{}
	var err error
	request := map[string]interface{}{
		"InstanceId": d.Id(),
	}

	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(client.GetRetryTimeout(d.Timeout(schema.TimeoutDelete)), func() *resource.RetryError {
		response, err = client.RpcPost("mse", "2019-05-31", action, nil, request, false)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, response, request)
		return nil
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
	}
	if fmt.Sprint(response["Success"]) == "false" {
		return WrapErrorf(fmt.Errorf("%s failed, response: %v", action, response), DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
	}
	stateConf := BuildStateConf([]string{}, []string{"DESTROY_SUCCESS"}, d.Timeout(schema.TimeoutDelete), 60*time.Second, mseService.MseClusterStateRefreshFunc(d.Id(), []string{"DESTROY_FAILED"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}
	return nil
}

func customizeMseClusterDiff(d *schema.ResourceDiff, meta interface{}) error {
	if d.Id() == "" {
		log.Printf("[DEBUG] This is a new resource")
		return nil
	}

	// 调用QueryClusterInfo接口，判断实例是否能够升级
	client := meta.(*connectivity.AliyunClient)

	queryClusterInfoRequest := map[string]interface{}{
		"InstanceId": d.Id(),
	}
	response, err := client.RpcPost("mse", "2019-05-31", "QueryClusterInfo", nil, queryClusterInfoRequest, false)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "QueryClusterInfo", AlibabaCloudSdkGoERROR)
	}
	log.Printf("[DEBUG] QueryClusterInfo response: %#v", response)
	data, ok := response["Data"].(map[string]interface{})
	if !ok {
		return WrapErrorf(fmt.Errorf("failed to get data from response: %v", response), DefaultErrorMsg, d.Id(), "QueryClusterInfo", AlibabaCloudSdkGoERROR)
	}

	canUpdate, ok := data["CanUpdate"].(bool)
	if !ok {
		return WrapErrorf(fmt.Errorf("failed to get canUpdate: %v", data), DefaultErrorMsg, d.Id(), "QueryClusterInfo", AlibabaCloudSdkGoERROR)
	}
	if !canUpdate {
		oldValue, _ := d.GetChange("version_code")
		d.SetNew("version_code", oldValue)
		return nil
	}

	// 只处理已存在的资源且 version_code 发生变化或为 LATEST 的情况
	if d.Id() != "" && (d.HasChange("version_code") || d.Get("version_code").(string) == "LATEST") {
		client := meta.(*connectivity.AliyunClient)
		mseService := MseService{client}

		// 获取当前实例信息
		object, err := mseService.DescribeMseCluster(d.Id())
		if err != nil {
			return WrapError(err)
		}

		log.Printf("[DEBUG] DescribeMseCluster response: %#v", object)

		versionCode, ok := object["ClusterVersion"].(string)
		if !ok || versionCode == "" {
			return WrapErrorf(fmt.Errorf("failed to get current version, cluster info: %v", object), DefaultErrorMsg, d.Id(), "GetClusterInfo", AlibabaCloudSdkGoERROR)
		}
		log.Printf("[DEBUG] Current VersionCode: %s", versionCode)

		getImageRequest := map[string]interface{}{
			"VersionCode": versionCode,
		}
		log.Printf("[DEBUG] GetImage request: %#v", getImageRequest)

		response, err := client.RpcPost("mse", "2019-05-31", "GetImage", nil, getImageRequest, false)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "GetImage", AlibabaCloudSdkGoERROR)
		}
		log.Printf("[DEBUG] GetImage response: %#v", response)

		data, ok := response["Data"].(map[string]interface{})
		if !ok {
			return WrapErrorf(fmt.Errorf("failed to get data from response: %v", response), DefaultErrorMsg, d.Id(), "GetImage", AlibabaCloudSdkGoERROR)
		}

		maxVersionCode, ok := data["MaxVersionCode"].(string)
		if !ok {
			return WrapErrorf(fmt.Errorf("failed to get maxVersionCode: %v", data), DefaultErrorMsg, d.Id(), "GetImage", AlibabaCloudSdkGoERROR)
		}
		log.Printf("[DEBUG] MaxVersionCode: %s", maxVersionCode)

		targetVersion := d.Get("version_code").(string)
		if targetVersion == "LATEST" {
			// 如果是 LATEST，设置为最新版本
			d.SetNew("version_code", maxVersionCode)
		} else if d.HasChange("version_code") {
			// 如果指定了具体版本且发生变化，验证是否为最新版本
			if targetVersion != maxVersionCode {
				return WrapErrorf(fmt.Errorf("can only upgrade to the latest version %s, but got %s. "+
					"You can also set version_code to 'LATEST' to always upgrade to the latest version",
					maxVersionCode, targetVersion), DefaultErrorMsg, d.Id(), "ValidateVersion", AlibabaCloudSdkGoERROR)
			}
		}
	}
	return nil
}

func convertMseChargeTypeToPaymentType(source interface{}) interface{} {
	switch source {
	case "POSTPAY":
		return "PayAsYouGo"
	case "PREPAY":
		return "Subscription"
	}
	return source
}

func convertMsePaymentTypeToChargeType(source interface{}) interface{} {
	switch source {
	case "PayAsYouGo":
		return "POSTPAY"
	case "Subscription":
		return "PREPAY"
	}
	return source
}
