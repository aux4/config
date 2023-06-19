const { ConfigLoader } = require("../../index");
const { Output: out } = require("@aux4/engine");
const ConfigSaver = require("../../lib/ConfigSaver");

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

  const config = ConfigLoader.load(await args.file);
  const response = config.set(name, value);

  const file = await args.file;

  if (!file) {
    out.println("file must be specified to save merged configuration".red);
    process.exit(1);
  }

  await ConfigSaver.save(file, config.config);

  if (response === undefined) {
    process.stdout.write("");
  } else if (typeof response === "object") {
    process.stdout.write(JSON.stringify(response, null, 2));
  } else {
    process.stdout.write(response.toString());
  }
}

module.exports = { setConfigExecutor };
