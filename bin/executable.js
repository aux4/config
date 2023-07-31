#!/usr/bin/env node

const { Engine, Config } = require("@aux4/engine");
const { getConfigExecutor } = require("./command/GetConfigExecutor");
const { mergeConfigExecutor } = require("./command/MergeConfigExecutor");
const { setConfigExecutor } = require("./command/SetConfigExecutor");

const args = process.argv.slice(2);

const config = new Config();
config.load({
  profiles: [
    {
      name: "main",
      commands: [
        {
          name: "get",
          execute: getConfigExecutor,
          help: {
            text: "Get a configuration value",
            variables: [
              {
                name: "file",
                text: "The configuration file to use",
                default: ""
              },
              {
                name: "name",
                text: "The name of the configuration value to get",
                default: ""
              }
            ]
          }
        },
        {
          name: "set",
          execute: setConfigExecutor,
          help: {
            text: "Set a configuration value",
            variables: [
              {
                name: "file",
                text: "The configuration file to use",
                default: ""
              },
              {
                name: "name",
                text: "The name of the configuration value to get"
              },
              {
                name: "value",
                text: "The value to set",
                default: ""
              }
            ]
          }
        },
        {
          name: "merge",
          execute: mergeConfigExecutor,
          help: {
            text: "Merge a configuration file into the current configuration",
            variables: [
              {
                name: "file",
                text: "The configuration file to merge",
                default: ""
              },
              {
                name: "save",
                text: "save the merged configuration to the specified file",
                default: false
              }
            ]
          }
        }
      ]
    }
  ]
});

(async () => {
  const engine = new Engine({ config });
  await engine.run(args);
})();
