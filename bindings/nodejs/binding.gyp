{
  "targets": [
    {
      "target_name": "gzh_manager_native",
      "sources": [
        "src/native/addon.cc"
      ],
      "include_dirs": [
        "<!@(node -p \"require('node-addon-api').include\")",
        "."
      ],
      "dependencies": [
        "<!(node -p \"require('node-addon-api').gyp\")"
      ],
      "libraries": [
        "-L<(module_root_dir)",
        "-lgzh_node"
      ],
      "cflags!": ["-fno-exceptions"],
      "cflags_cc!": ["-fno-exceptions"],
      "xcode_settings": {
        "GCC_ENABLE_CPP_EXCEPTIONS": "YES",
        "CLANG_CXX_LIBRARY": "libc++",
        "MACOSX_DEPLOYMENT_TARGET": "10.7"
      },
      "msvs_settings": {
        "VCCLCompilerTool": {
          "ExceptionHandling": 1
        }
      },
      "conditions": [
        [
          "OS=='mac'",
          {
            "libraries": [
              "-Wl,-rpath,@loader_path"
            ]
          }
        ],
        [
          "OS=='linux'",
          {
            "libraries": [
              "-Wl,-rpath,$ORIGIN"
            ]
          }
        ]
      ]
    }
  ]
}