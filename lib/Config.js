const fs = require("fs");
const yaml = require("js-yaml");
const { recursive } = require("merge");
const { isJsonFile } = require("./Utils");

class Config {
  constructor(file) {
    if (typeof file === "object") {
      this.config = adaptConfig(file);
      return;
    }

    if (!fs.existsSync(file)) throw new Error(`Config file ${file} not found`);

    const fileContent = fs.readFileSync(file, { encoding: "utf8" });
    const isJson = isJsonFile(file, fileContent);

    if (isJson) {
      try {
        this.config = adaptConfig(JSON.parse(fileContent));
      } catch (e) {
        throw new Error(`Error parsing JSON config file ${file}: ${e.message}`);
      }
    } else {
      try {
        this.config = adaptConfig(yaml.load(fileContent, "utf8"));
      } catch (e) {
        throw new Error(`Error parsing YAML config file ${file}: ${e.message}`);
      }
    }
  }

  merge(newConfig) {
    return recursive(this.config, { ...newConfig.config });
  }

  get(name) {
    let value = this.config.config;
    if (name) {
      const configPath = name.split("/");
      for (const path of configPath) {
        if (path === "") continue;

        if (value[path] === undefined) {
          return undefined;
        }
        value = value[path];
      }
    }

    return value;
  }

  set(name, value) {
    let config = this.config.config;
    const configPath = name.split("/");
    for (const path of configPath.slice(0, -1)) {
      if (path === "") continue;

      if (config[path] === undefined) {
        config[path] = {};
      }
      config = config[path];
    }
    config[configPath[configPath.length - 1]] = value;
    return value;
  }
}

function adaptConfig(content = {}) {
  if (content.config) {
    return content;
  }
  return { config: content };
}

module.exports = Config;
