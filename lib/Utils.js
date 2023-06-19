const path = require("path");
const fs = require("fs");

const CONFIG_FILE_PATTERN = ["config.json", "config.yaml", "config.yml"];

function isJsonFile(file, fileContent = "") {
  return (
    file.toLowerCase().endsWith(".json") || file.toLowerCase().endsWith(".aux4") || fileContent.trim().startsWith("{")
  );
}

function findConfigFile(folder = ".", folders = []) {
  const fullPath = path.resolve(folder);

  for (const configFilePattern of CONFIG_FILE_PATTERN) {
    const configFile = path.join(fullPath, configFilePattern);
    if (fs.existsSync(configFile)) {
      folders.unshift(configFile);
      break;
    }
  }

  const parentFolder = path.resolve(path.join(fullPath, ".."));
  if (parentFolder !== fullPath) {
    return findConfigFile(parentFolder, folders);
  }

  return folders;
}

function readStdIn() {
  return new Promise(resolve => {
    let inputString = "";

    const stdin = process.openStdin();

    stdin.on("data", function (data) {
      inputString += data;
    });

    stdin.on("end", function () {
      resolve(inputString);
    });
  });
}

module.exports = { isJsonFile, findConfigFile, readStdIn };
