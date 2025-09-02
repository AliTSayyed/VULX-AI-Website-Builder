from fastapi import FastAPI

ai_service = FastAPI()

@ai_service.get('/')
def greeting():
    return {"message":"hello"}    

