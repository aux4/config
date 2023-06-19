const { ConfigLoader } = require("../../index");

async function getConfigExecutor(args) {
  const config = ConfigLoader.load(await args.file);
  const response = config.get(await args.name);

  if (response === undefined) {
    process.stdout.write("");
  } else if (typeof response === "object") {
    process.stdout.write(JSON.stringify(response, null, 2));
  } else {
    process.stdout.write(response.toString());
  }
}

module.exports = { getConfigExecutor };
