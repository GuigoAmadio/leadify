import os
from dotenv import load_dotenv
from sentence_transformers import SentenceTransformer

load_dotenv()

def cache_embedding_model():
    model_name = os.getenv("EMBEDDING_MODEL_NAME")

    if not model_name:
        print("[ERRO] Nao foi possivel carregar o nome do modelo de embedding o .env")
        return
    
    print(f"[*] Inicializando download e cache para: {model_name}")

    try:
        model = SentenceTransformer(model_name, trust_remote_code=True)
        print(f"{model_name} cacheado com sucesso")
        print(f"Max SL: {model.max_seq_length} tokens")
        print(f"Base Dimensions: {model.get_sentence_embedding_dimension}")

    except Exception as e:
        print(f"\n Erro: {e}")

if __name__ == "__main__":
    cache_embedding_model()