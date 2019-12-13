package context

import (
	"reflect"
	"testing"
)

func Test_Add(t *testing.T) {
	rawResource := []byte(`
	{
		"apiVersion": "v1",
		"kind": "Pod",
		"metadata": {
		   "name": "image-with-hostpath",
		   "labels": {
			  "app.type": "prod",
			  "namespace": "my-namespace"
		   }
		},
		"spec": {
		   "containers": [
			  {
				 "name": "image-with-hostpath",
				 "image": "docker.io/nautiker/curl",
				 "volumeMounts": [
					{
					   "name": "var-lib-etcd",
					   "mountPath": "/var/lib"
					}
				 ]
			  }
		   ],
		   "volumes": [
			  {
				 "name": "var-lib-etcd",
				 "emptyDir": {}
			  }
		   ]
		}
	 }
			`)

	expectedResult := "my-namespace"

	var err error
	ctx := NewContext()
	ctx.Add("resource", rawResource)
	query := "resource.metadata.labels.namespace"
	result, err := ctx.Query(query)
	if err != nil {
		t.Error(err)
	}
	t.Log(expectedResult)
	t.Log(result)
	if !reflect.DeepEqual(expectedResult, result) {
		t.Error("exected result does not match")
	}
}

func Test_AddUser(t *testing.T) {
	rawUser := []byte(`
	{
		"userInfo": {
				"username": "admin",
				"uid": "014fbff9a07c",
				"groups": ["system:authenticated","my-admin-group"],
				"extra": {
					"some-key":["some-value1", "some-value2"]
				}
			}
	}
	`)
	expectedResult := "admin"

	var err error
	ctx := NewContext()
	ctx.Add("user", rawUser)
	query := "user.userInfo.username"
	result, err := ctx.Query(query)
	if err != nil {
		t.Error(err)
	}
	t.Log(expectedResult)
	t.Log(result)
	if !reflect.DeepEqual(expectedResult, result) {
		t.Error("exected result does not match")
	}
}