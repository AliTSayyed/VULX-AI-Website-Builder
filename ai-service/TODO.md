some folder structure like so
ai_service/
├── main.py
├── utils/
│ ├── **init**.py
│ ├── logging.py
│ └── exceptions.py
├── clients/
│ ├── **init**.py
│ ├── ai_client.py
│ └── e2b_client.py
├── api/
│ ├── **init**.py
│ ├── routes/
│ │ ├── **init**.py
│ │ ├── agents.py
│ │ └── sandbox.py
│ └── schemas/
│ ├── **init**.py
│ ├── agent_schemas.py
│ └── sandbox_schemas.py
├── services/
│ ├── **init**.py
│ ├── agent_service.py
│ └── sandbox_service.py
└── config/
├── **init**.py
└── settings.py

Treat FastAPI service as a "package" or library that Go API consumes via HTTP calls.
