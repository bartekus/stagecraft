// SPDX-License-Identifier: AGPL-3.0-or-later
//
// Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.
//
// Copyright (C) 2025  Bartek Kus
//
// This program is free software licensed under the terms of the GNU AGPL v3 or later.
//
// See https://www.gnu.org/licenses/ for license details.
//

// Basic Node.js backend example
const express = require('express');
const app = express();
const port = process.env.PORT || 4000;

app.get('/', (req, res) => {
  res.json({ message: 'Hello from Stagecraft generic provider!' });
});

app.get('/health', (req, res) => {
  res.json({ status: 'ok' });
});

app.listen(port, '0.0.0.0', () => {
  console.log(`Server running on http://0.0.0.0:${port}`);
});

