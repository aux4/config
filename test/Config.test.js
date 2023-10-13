const { Config } = require("..");

describe("Config", () => {
  let config, value;

  describe("when file is json", () => {
    beforeEach(() => {
      config = new Config("./test/config.json");
    });

    describe("get dev", () => {
      beforeEach(() => {
        value = config.get("dev");
      });

      it("should return dev value", () => {
        expect(value).toEqual({
          host: "localhost",
          port: 3000
        });
      });
    });

    describe("get dev/host", () => {
      beforeEach(() => {
        value = config.get("dev/host");
      });

      it("should return dev host value", () => {
        expect(value).toEqual("localhost");
      });
    });

    describe("get undefined value", () => {
      beforeEach(() => {
        value = config.get("dev/host/other");
      });

      it("should return undefined", () => {
        expect(value).toBeUndefined();
      });
    });
  });

  describe("when file is yaml", () => {
    beforeEach(() => {
      config = new Config("./test/config.yaml");
    });

    describe("get dev", () => {
      beforeEach(() => {
        value = config.get("dev");
      });

      it("should return dev value", () => {
        expect(value).toEqual({
          host: "localhost",
          port: 3000
        });
      });
    });
  });

  describe("when file is wrong json", () => {
    it("should throw error", () => {
      expect(() => {
        new Config("./test/wrong.json");
      }).toThrow(
        "Error parsing JSON config file ./test/wrong.json: Expected property name or '}' in JSON at position 4"
      );
    });
  });

  describe("when file is wrong yaml", () => {
    it("should throw error", () => {
      expect(() => {
        new Config("./test/wrong.yaml");
      }).toThrow("Error parsing YAML config file ./test/wrong.yaml: bad indentation of a mapping entry");
    });
  });

  describe("when file does not exist", () => {
    it("should throw error", () => {
      expect(() => {
        new Config("./test/do-not-exist.yaml");
      }).toThrow("Config file ./test/do-not-exist.yaml not found");
    });
  });
});
