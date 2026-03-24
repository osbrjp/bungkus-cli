import { Project } from "ts-morph";
import fs from "fs"
import path from "path"

const configPath = path.resolve("astro.config.mjs")

const patch = () => {};

const readConfigFile = () => {
  if (!fs.existsSync(configPath)) {
    console.log("astro.config.mjs not found")
    process.exit(0)
  }
};
