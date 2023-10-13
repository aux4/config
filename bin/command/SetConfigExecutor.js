const { ConfigLoader } = require("../../index");
const { Output: out } = require("@aux4/engine");
const ConfigSaver = require("../../lib/ConfigSaver");
const { findConfigFile } = require("../../lib/Utils");

async function setConfigExecutor(args) {
  let name = await args.name;
  let value = await args.value;

  if (name.includes("=")) {
    const parts = name.split("=");
    name = parts[0];
    value = parts[1];
  }

  if (name === undefined) {
    out.println("name is required".red);
    process.exit(1);
  }

  if (value === undefined) {
    out.println("value is required".red);
    process.exit(1);
  }

  const file = await args.file;

  const config = ConfigLoader.load(file);
  config.set(name, value);

  await ConfigSaver.save(file || "config.yaml", config.config);
}

module.exports = { setConfigExecutor };
