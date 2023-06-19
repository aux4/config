const fs = require("fs");
const yaml = require("js-yaml");
const { isJsonFile } = require("./Utils");

class ConfigSaver {
  static async save(file, config) {
    const isJson = isJsonFile(file);

    let content;
    if (isJson) {
      content = JSON.stringify(config, null, 2);
    } else {
      content = yaml.dump(config, file);
    }

    try {
      fs.writeFileSync(file, content);
    } catch (e) {
      throw new Error(`Error saving config file: ${e.message}`);
    }
  }
}

module.exports = ConfigSaver;
