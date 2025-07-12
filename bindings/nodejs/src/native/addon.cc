#include <napi.h>
#include <cstdlib>
#include <cstring>

extern "C" {
    // C function declarations from Go library
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

    int gzh_node_client_create(gzh_client_config_t* config);
    void gzh_node_client_destroy(int clientID);
    gzh_result_t* gzh_node_bulk_clone(int clientID, gzh_bulk_clone_request_t* request);
    gzh_result_t* gzh_node_list_plugins(int clientID);
    gzh_result_t* gzh_node_execute_plugin(int clientID, char* plugin_name, char* method, char* args_json, int timeout_seconds);
    gzh_result_t* gzh_node_health(int clientID);
    void gzh_node_free_result(gzh_result_t* result);
    void gzh_node_free_string(char* str);
}

// Helper function to convert Napi::Object to gzh_client_config_t
gzh_client_config_t* ObjectToClientConfig(const Napi::Object& obj) {
    gzh_client_config_t* config = new gzh_client_config_t();
    memset(config, 0, sizeof(gzh_client_config_t));

    if (obj.Has("timeout")) {
        config->timeout = obj.Get("timeout").As<Napi::Number>().Int64Value();
    }
    if (obj.Has("retryCount")) {
        config->retry_count = obj.Get("retryCount").As<Napi::Number>().Int32Value();
    }
    if (obj.Has("enablePlugins")) {
        config->enable_plugins = obj.Get("enablePlugins").As<Napi::Boolean>().Value() ? 1 : 0;
    }
    if (obj.Has("pluginDir")) {
        std::string pluginDir = obj.Get("pluginDir").As<Napi::String>().Utf8Value();
        config->plugin_dir = strdup(pluginDir.c_str());
    }
    if (obj.Has("logLevel")) {
        std::string logLevel = obj.Get("logLevel").As<Napi::String>().Utf8Value();
        config->log_level = strdup(logLevel.c_str());
    }
    if (obj.Has("logFile")) {
        std::string logFile = obj.Get("logFile").As<Napi::String>().Utf8Value();
        config->log_file = strdup(logFile.c_str());
    }

    return config;
}

// Helper function to free gzh_client_config_t
void FreeClientConfig(gzh_client_config_t* config) {
    if (config) {
        if (config->plugin_dir) free(config->plugin_dir);
        if (config->log_level) free(config->log_level);
        if (config->log_file) free(config->log_file);
        delete config;
    }
}

// Helper function to convert gzh_result_t to Napi::Object
Napi::Object ResultToObject(Napi::Env env, gzh_result_t* result) {
    Napi::Object obj = Napi::Object::New(env);
    
    obj.Set("success", Napi::Boolean::New(env, result->success == 1));
    
    if (result->error_msg) {
        obj.Set("error", Napi::String::New(env, result->error_msg));
    } else {
        obj.Set("error", env.Null());
    }
    
    if (result->data_json) {
        obj.Set("data", Napi::String::New(env, result->data_json));
    } else {
        obj.Set("data", env.Null());
    }
    
    return obj;
}

// Create client
Napi::Value CreateClient(const Napi::CallbackInfo& info) {
    Napi::Env env = info.Env();
    
    gzh_client_config_t* config = nullptr;
    if (info.Length() > 0 && info[0].IsObject()) {
        config = ObjectToClientConfig(info[0].As<Napi::Object>());
    }
    
    int clientID = gzh_node_client_create(config);
    
    if (config) {
        FreeClientConfig(config);
    }
    
    if (clientID < 0) {
        Napi::TypeError::New(env, "Failed to create client").ThrowAsJavaScriptException();
        return env.Null();
    }
    
    return Napi::Number::New(env, clientID);
}

// Destroy client
Napi::Value DestroyClient(const Napi::CallbackInfo& info) {
    Napi::Env env = info.Env();
    
    if (info.Length() < 1 || !info[0].IsNumber()) {
        Napi::TypeError::New(env, "Expected client ID").ThrowAsJavaScriptException();
        return env.Undefined();
    }
    
    int clientID = info[0].As<Napi::Number>().Int32Value();
    gzh_node_client_destroy(clientID);
    
    return env.Undefined();
}

// Bulk clone
Napi::Value BulkClone(const Napi::CallbackInfo& info) {
    Napi::Env env = info.Env();
    
    if (info.Length() < 2 || !info[0].IsNumber() || !info[1].IsObject()) {
        Napi::TypeError::New(env, "Expected client ID and request object").ThrowAsJavaScriptException();
        return env.Null();
    }
    
    int clientID = info[0].As<Napi::Number>().Int32Value();
    Napi::Object requestObj = info[1].As<Napi::Object>();
    
    gzh_bulk_clone_request_t request;
    memset(&request, 0, sizeof(request));
    
    // Convert request object to C struct
    if (requestObj.Has("platforms")) {
        std::string platforms = requestObj.Get("platforms").As<Napi::String>().Utf8Value();
        request.platforms_json = strdup(platforms.c_str());
    }
    if (requestObj.Has("outputDir")) {
        std::string outputDir = requestObj.Get("outputDir").As<Napi::String>().Utf8Value();
        request.output_dir = strdup(outputDir.c_str());
    }
    if (requestObj.Has("concurrency")) {
        request.concurrency = requestObj.Get("concurrency").As<Napi::Number>().Int32Value();
    }
    if (requestObj.Has("strategy")) {
        std::string strategy = requestObj.Get("strategy").As<Napi::String>().Utf8Value();
        request.strategy = strdup(strategy.c_str());
    }
    if (requestObj.Has("includePrivate")) {
        request.include_private = requestObj.Get("includePrivate").As<Napi::Boolean>().Value() ? 1 : 0;
    }
    if (requestObj.Has("filters")) {
        std::string filters = requestObj.Get("filters").As<Napi::String>().Utf8Value();
        request.filters_json = strdup(filters.c_str());
    }
    
    gzh_result_t* result = gzh_node_bulk_clone(clientID, &request);
    
    // Free request strings
    if (request.platforms_json) free(request.platforms_json);
    if (request.output_dir) free(request.output_dir);
    if (request.strategy) free(request.strategy);
    if (request.filters_json) free(request.filters_json);
    
    Napi::Object resultObj = ResultToObject(env, result);
    gzh_node_free_result(result);
    
    return resultObj;
}

// List plugins
Napi::Value ListPlugins(const Napi::CallbackInfo& info) {
    Napi::Env env = info.Env();
    
    if (info.Length() < 1 || !info[0].IsNumber()) {
        Napi::TypeError::New(env, "Expected client ID").ThrowAsJavaScriptException();
        return env.Null();
    }
    
    int clientID = info[0].As<Napi::Number>().Int32Value();
    gzh_result_t* result = gzh_node_list_plugins(clientID);
    
    Napi::Object resultObj = ResultToObject(env, result);
    gzh_node_free_result(result);
    
    return resultObj;
}

// Execute plugin
Napi::Value ExecutePlugin(const Napi::CallbackInfo& info) {
    Napi::Env env = info.Env();
    
    if (info.Length() < 4 || !info[0].IsNumber() || !info[1].IsString() || 
        !info[2].IsString() || !info[3].IsString()) {
        Napi::TypeError::New(env, "Expected client ID, plugin name, method, and args").ThrowAsJavaScriptException();
        return env.Null();
    }
    
    int clientID = info[0].As<Napi::Number>().Int32Value();
    std::string pluginName = info[1].As<Napi::String>().Utf8Value();
    std::string method = info[2].As<Napi::String>().Utf8Value();
    std::string argsJson = info[3].As<Napi::String>().Utf8Value();
    int timeout = info.Length() > 4 ? info[4].As<Napi::Number>().Int32Value() : 30;
    
    gzh_result_t* result = gzh_node_execute_plugin(
        clientID, 
        const_cast<char*>(pluginName.c_str()),
        const_cast<char*>(method.c_str()),
        const_cast<char*>(argsJson.c_str()),
        timeout
    );
    
    Napi::Object resultObj = ResultToObject(env, result);
    gzh_node_free_result(result);
    
    return resultObj;
}

// Health check
Napi::Value Health(const Napi::CallbackInfo& info) {
    Napi::Env env = info.Env();
    
    if (info.Length() < 1 || !info[0].IsNumber()) {
        Napi::TypeError::New(env, "Expected client ID").ThrowAsJavaScriptException();
        return env.Null();
    }
    
    int clientID = info[0].As<Napi::Number>().Int32Value();
    gzh_result_t* result = gzh_node_health(clientID);
    
    Napi::Object resultObj = ResultToObject(env, result);
    gzh_node_free_result(result);
    
    return resultObj;
}

// Initialize the addon
Napi::Object Init(Napi::Env env, Napi::Object exports) {
    exports.Set(Napi::String::New(env, "createClient"), Napi::Function::New(env, CreateClient));
    exports.Set(Napi::String::New(env, "destroyClient"), Napi::Function::New(env, DestroyClient));
    exports.Set(Napi::String::New(env, "bulkClone"), Napi::Function::New(env, BulkClone));
    exports.Set(Napi::String::New(env, "listPlugins"), Napi::Function::New(env, ListPlugins));
    exports.Set(Napi::String::New(env, "executePlugin"), Napi::Function::New(env, ExecutePlugin));
    exports.Set(Napi::String::New(env, "health"), Napi::Function::New(env, Health));
    
    return exports;
}

NODE_API_MODULE(gzh_manager_native, Init)