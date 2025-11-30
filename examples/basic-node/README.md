# Basic Node.js Example

This example demonstrates Stagecraft with a generic Node.js backend and raw SQL migrations.

## Features

- **Generic Backend Provider**: Uses the `generic` provider to run a Node.js Express server
- **Raw Migration Engine**: Uses the `raw` migration engine for SQL migrations
- **Simple Configuration**: Minimal `stagecraft.yml` showing provider-scoped config

## Structure

```
basic-node/
├── stagecraft.yml          # Stagecraft configuration
├── backend/
│   ├── index.js            # Express server
│   ├── package.json        # Node.js dependencies
│   └── Dockerfile          # Docker build file
└── migrations/
    └── 001_initial.sql     # SQL migration file
```

## Configuration

The `stagecraft.yml` shows:

1. **Generic Backend Provider**:
   ```yaml
   backend:
     provider: generic
     providers:
       generic:
         dev:
           command: ["npm", "run", "dev"]
           workdir: "./backend"
   ```

2. **Raw Migration Engine**:
   ```yaml
   databases:
     main:
       migrations:
         engine: raw
         path: ./migrations
   ```

## Usage

1. Install dependencies:
   ```bash
   cd backend
   npm install
   ```

2. Run development server:
   ```bash
   stagecraft dev
   ```

3. The generic provider will execute `npm run dev` in the `./backend` directory.

## Notes

- This example uses the generic provider, which is framework-agnostic
- The raw migration engine reads SQL files from the migrations directory
- No Encore.ts or other framework-specific tools required

