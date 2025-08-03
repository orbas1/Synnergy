import { Router } from "express";
import * as ctrl from "../controllers/contractController.js";

const router = Router();

router.get("/", ctrl.list);
router.post("/", ctrl.deploy);
router.get("/:id", ctrl.get);
router.delete("/:id", ctrl.remove);
router.get("/:id/wasm", ctrl.wasm);

export default router;
