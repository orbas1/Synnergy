require("dotenv").config();
const express = require("express");
const { exec } = require("child_process");

const app = express();
app.use(express.json());

const CLI = process.env.CLI_PATH || "synnergy";
const MARKETPLACE_ADDRESS = process.env.MARKETPLACE_ADDRESS || "";

const services = [
  {
    id: "1",
    name: "Image Recognition",
    price: 1,
    description: "Identify objects in images using AI.",
  },
  {
    id: "2",
    name: "Text Summarization",
    price: 1,
    description: "Generate concise summaries from text.",
  },
  {
    id: "3",
    name: "Voice Generation",
    price: 1,
    description: "Convert text to human-like speech.",
  },
];

app.get("/api/services", (req, res) => {
  res.json(services);
});

app.post("/api/purchase", (req, res) => {
  const { id } = req.body;
  const svc = services.find((s) => s.id === id);
  if (!svc) return res.status(404).json({ error: "Service not found" });

  if (!MARKETPLACE_ADDRESS) {
    return res.status(500).json({ error: "contract address not configured" });
  }

  const cmd = `${CLI} contracts invoke ${MARKETPLACE_ADDRESS} --method buyService --args ${id} --gas 200000`;
  exec(cmd, (err, stdout, stderr) => {
    if (err) {
      console.error(stderr);
      return res.status(500).json({ error: "purchase failed" });
    }
    console.log(`Purchased service ${id}`);
    res.json({ output: stdout.trim() });
  });
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => console.log(`AI Marketplace server running on ${PORT}`));
