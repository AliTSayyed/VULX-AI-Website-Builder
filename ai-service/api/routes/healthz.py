from fastapi import APIRouter

router = APIRouter()

@router.get('/healthz', tags=["Healthz"])
def healthz():
    return {"status":"ok"}
