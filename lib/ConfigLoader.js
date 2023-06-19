const Config = require("./Config");
const { findConfigFile } = require("./Utils");

class ConfigLoader {
  static load(file) {
    let configurationFiles;

    if (file) {
      configurationFiles = [file];
    } else {
      configurationFiles = findConfigFile();
    }

    if (configurationFiles.length === 0) {
      console.error("No configuration file found");
      process.exit(1);
    }

    const config = new Config(configurationFiles[0]);

    if (configurationFiles.length > 1) {
      configurationFiles.slice(1).forEach(configFile => {
        const nextConfig = new Config(configFile);
        config.merge(nextConfig);
      });
    }

    return config;
  }
}

module.exports = ConfigLoader;
