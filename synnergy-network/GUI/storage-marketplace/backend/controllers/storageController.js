const service = require("../services/storageService");

exports.pin = async (req, res, next) => {
  try {
    const file = await service.pin(req.body.cid, req.body.meta || {});
    res.status(201).json(file);
  } catch (err) {
    next(err);
  }
};

exports.listPins = async (req, res, next) => {
  try {
    const files = await service.listPins();
    res.json(files);
  } catch (err) {
    next(err);
  }
};

exports.retrieve = async (req, res, next) => {
  try {
    const file = await service.retrieve(req.params.cid);
    if (!file) return res.status(404).json({ error: "Not found" });
    res.json(file);
  } catch (err) {
    next(err);
  }
};

exports.exists = async (req, res, next) => {
  try {
    const ok = await service.exists(req.params.cid);
    res.json({ exists: ok });
  } catch (err) {
    next(err);
  }
};

exports.createStorage = async (req, res, next) => {
  try {
    const storage = await service.createStorage(req.body);
    res.status(201).json(storage);
  } catch (err) {
    next(err);
  }
};

exports.listStorages = async (req, res, next) => {
  try {
    const storages = await service.listStorages();
    res.json(storages);
  } catch (err) {
    next(err);
  }
};
