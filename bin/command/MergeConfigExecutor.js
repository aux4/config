const colors = require("colors");
const { ConfigLoader, Config } = require("../..");
const { Output: out } = require("@aux4/engine");
const { readStdIn } = require("../../lib/Utils");
const ConfigSaver = require("../../lib/ConfigSaver");

async function mergeConfigExecutor(args) {
  if (process.stdin.isTTY) {
    throw new Error("No input configuration provided");
  }

  const file = await args.file;
  const config = ConfigLoader.load(file);

  const inputConfig = await readStdIn();

  try {
    const previousConfigObject = JSON.parse(inputConfig);

    const configuration = previousConfigObject.config ? { config: previousConfigObject } : previousConfigObject;

    const previousConfig = new Config(configuration);
    const merged = config.merge(previousConfig);

    process.stdout.write(JSON.stringify(merged.config, null, 2));

    if (await args.save) {
      if (!file) {
        out.println("file must be specified to save merged configuration".red);
        process.exit(1);
      }

      await ConfigSaver.save(file, merged);
    }
  } catch (e) {
    out.println(`Error parsing input configuration: ${e.message}`, e);
    process.exit(1);
  }
}

module.exports = { mergeConfigExecutor };
