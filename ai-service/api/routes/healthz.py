from main import ai_service

@ai_service.get('/healthz')
def healthz():
    return {"status":"ok"}