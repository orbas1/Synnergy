const express = require('express');
const app = express();
app.use(express.json());

const services = [
  { id: '1', name: 'Image Recognition', description: 'Identify objects in images using AI.' },
  { id: '2', name: 'Text Summarization', description: 'Generate concise summaries from text.' },
  { id: '3', name: 'Voice Generation', description: 'Convert text to human-like speech.' }
];

app.get('/api/services', (req, res) => {
  res.json(services);
});

app.post('/api/purchase', (req, res) => {
  const { id } = req.body;
  const svc = services.find(s => s.id === id);
  if (!svc) return res.status(404).json({ error: 'Service not found' });
  // In a real implementation this would trigger a smart contract call.
  console.log(`Purchased service ${id}`);
  res.json({ status: 'ok' });
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => console.log(`AI Marketplace server running on ${PORT}`));
