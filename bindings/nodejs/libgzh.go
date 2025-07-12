package main

/*
#include <stdlib.h>
#include <string.h>

typedef struct {
    int64_t timeout;
    int retry_count;
    int enable_plugins;
    char* plugin_dir;
    char* log_level;
    char* log_file;
} gzh_client_config_t;

typedef struct {
    char* platforms_json;
    char* output_dir;
    int concurrency;
    char* strategy;
    int include_private;
    char* filters_json;
} gzh_bulk_clone_request_t;

typedef struct {
    int success;
    char* error_msg;
    char* data_json;
} gzh_result_t;
*/
import "C"

import (
	"context"
	"encoding/json"
	"time"
	"unsafe"

	"github.com/gizzahub/gzh-manager-go/pkg/gzhclient"
)

var (
	clients = make(map[int]*gzhclient.Client)
	nextID  = 1
)

//export gzh_node_client_create
func gzh_node_client_create(config *C.gzh_client_config_t) C.int {
	clientConfig := gzhclient.DefaultConfig()

	if config != nil {
		if config.timeout > 0 {
			clientConfig.Timeout = time.Duration(config.timeout) * time.Second
		}
		if config.retry_count > 0 {
			clientConfig.RetryCount = int(config.retry_count)
		}
		clientConfig.EnablePlugins = config.enable_plugins == 1

		if config.plugin_dir != nil {
			clientConfig.PluginDir = C.GoString(config.plugin_dir)
		}
		if config.log_level != nil {
			clientConfig.LogLevel = C.GoString(config.log_level)
		}
		if config.log_file != nil {
			clientConfig.LogFile = C.GoString(config.log_file)
		}
	}

	client, err := gzhclient.NewClient(clientConfig)
	if err != nil {
		return -1
	}

	clientID := nextID
	nextID++
	clients[clientID] = client

	return C.int(clientID)
}

//export gzh_node_client_destroy
func gzh_node_client_destroy(clientID C.int) {
	id := int(clientID)
	if client, exists := clients[id]; exists {
		client.Close()
		delete(clients, id)
	}
}

//export gzh_node_bulk_clone
func gzh_node_bulk_clone(clientID C.int, request *C.gzh_bulk_clone_request_t) *C.gzh_result_t {
	id := int(clientID)
	client, exists := clients[id]
	if !exists {
		return &C.gzh_result_t{
			success:   0,
			error_msg: C.CString("Client not found"),
			data_json: nil,
		}
	}

	if request == nil {
		return &C.gzh_result_t{
			success:   0,
			error_msg: C.CString("Request is null"),
			data_json: nil,
		}
	}

	// Parse platforms JSON
	var platforms []gzhclient.PlatformConfig
	if request.platforms_json != nil {
		platformsStr := C.GoString(request.platforms_json)
		if err := json.Unmarshal([]byte(platformsStr), &platforms); err != nil {
			return &C.gzh_result_t{
				success:   0,
				error_msg: C.CString("Invalid platforms JSON: " + err.Error()),
				data_json: nil,
			}
		}
	}

	// Parse filters JSON
	var filters gzhclient.CloneFilters
	if request.filters_json != nil {
		filtersStr := C.GoString(request.filters_json)
		if err := json.Unmarshal([]byte(filtersStr), &filters); err != nil {
			return &C.gzh_result_t{
				success:   0,
				error_msg: C.CString("Invalid filters JSON: " + err.Error()),
				data_json: nil,
			}
		}
	}

	// Create bulk clone request
	bulkCloneReq := gzhclient.BulkCloneRequest{
		Platforms:      platforms,
		OutputDir:      C.GoString(request.output_dir),
		Concurrency:    int(request.concurrency),
		Strategy:       C.GoString(request.strategy),
		IncludePrivate: request.include_private == 1,
		Filters:        filters,
	}

	// Execute bulk clone
	result, err := client.BulkClone(context.Background(), bulkCloneReq)
	if err != nil {
		return &C.gzh_result_t{
			success:   0,
			error_msg: C.CString(err.Error()),
			data_json: nil,
		}
	}

	// Marshal result to JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return &C.gzh_result_t{
			success:   0,
			error_msg: C.CString(err.Error()),
			data_json: nil,
		}
	}

	return &C.gzh_result_t{
		success:   1,
		error_msg: nil,
		data_json: C.CString(string(resultJSON)),
	}
}

//export gzh_node_list_plugins
func gzh_node_list_plugins(clientID C.int) *C.gzh_result_t {
	id := int(clientID)
	client, exists := clients[id]
	if !exists {
		return &C.gzh_result_t{
			success:   0,
			error_msg: C.CString("Client not found"),
			data_json: nil,
		}
	}

	plugins, err := client.ListPlugins()
	if err != nil {
		return &C.gzh_result_t{
			success:   0,
			error_msg: C.CString(err.Error()),
			data_json: nil,
		}
	}

	pluginsJSON, err := json.Marshal(plugins)
	if err != nil {
		return &C.gzh_result_t{
			success:   0,
			error_msg: C.CString(err.Error()),
			data_json: nil,
		}
	}

	return &C.gzh_result_t{
		success:   1,
		error_msg: nil,
		data_json: C.CString(string(pluginsJSON)),
	}
}

//export gzh_node_execute_plugin
func gzh_node_execute_plugin(clientID C.int, plugin_name *C.char, method *C.char, args_json *C.char, timeout_seconds C.int) *C.gzh_result_t {
	id := int(clientID)
	client, exists := clients[id]
	if !exists {
		return &C.gzh_result_t{
			success:   0,
			error_msg: C.CString("Client not found"),
			data_json: nil,
		}
	}

	// Parse arguments JSON
	var args map[string]interface{}
	if args_json != nil {
		argsStr := C.GoString(args_json)
		if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
			return &C.gzh_result_t{
				success:   0,
				error_msg: C.CString("Invalid args JSON: " + err.Error()),
				data_json: nil,
			}
		}
	}

	// Create plugin execute request
	executeReq := gzhclient.PluginExecuteRequest{
		PluginName: C.GoString(plugin_name),
		Method:     C.GoString(method),
		Args:       args,
		Timeout:    time.Duration(timeout_seconds) * time.Second,
	}

	// Execute plugin
	result, err := client.ExecutePlugin(context.Background(), executeReq)
	if err != nil {
		return &C.gzh_result_t{
			success:   0,
			error_msg: C.CString(err.Error()),
			data_json: nil,
		}
	}

	// Marshal result to JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return &C.gzh_result_t{
			success:   0,
			error_msg: C.CString(err.Error()),
			data_json: nil,
		}
	}

	return &C.gzh_result_t{
		success:   1,
		error_msg: nil,
		data_json: C.CString(string(resultJSON)),
	}
}

//export gzh_node_health
func gzh_node_health(clientID C.int) *C.gzh_result_t {
	id := int(clientID)
	client, exists := clients[id]
	if !exists {
		return &C.gzh_result_t{
			success:   0,
			error_msg: C.CString("Client not found"),
			data_json: nil,
		}
	}

	health := client.Health()
	healthJSON, err := json.Marshal(health)
	if err != nil {
		return &C.gzh_result_t{
			success:   0,
			error_msg: C.CString(err.Error()),
			data_json: nil,
		}
	}

	return &C.gzh_result_t{
		success:   1,
		error_msg: nil,
		data_json: C.CString(string(healthJSON)),
	}
}

//export gzh_node_free_result
func gzh_node_free_result(result *C.gzh_result_t) {
	if result != nil {
		if result.error_msg != nil {
			C.free(unsafe.Pointer(result.error_msg))
		}
		if result.data_json != nil {
			C.free(unsafe.Pointer(result.data_json))
		}
		C.free(unsafe.Pointer(result))
	}
}

//export gzh_node_free_string
func gzh_node_free_string(str *C.char) {
	if str != nil {
		C.free(unsafe.Pointer(str))
	}
}

func main() {
	// This is required for building as a Node.js addon
}
