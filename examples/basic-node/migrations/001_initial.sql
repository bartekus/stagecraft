-- SPDX-License-Identifier: AGPL-3.0-or-later
--
-- Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.
--
-- Copyright (C) 2025  Bartek Kus
--
-- This program is free software licensed under the terms of the GNU AGPL v3 or later.
--
-- See https://www.gnu.org/licenses/ for license details.
--

-- Initial database schema
-- This is an example migration file for the raw migration engine

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

