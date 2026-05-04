import os
import json
import psycopg2
from pgvector.psycopg2 import register_vector
from dotenv import load_dotenv
from sentence_transformers import SentenceTransformer

load_dotenv()

class ProfileEmbedder:
    def __init__(self):
        self.db_url = os.getenv("DATABASE_URL")
        model_name = os.getenv("EMBEDDING_MODEL_NAME", "google/embeddinggemma-300")

        print(f"[*] Carregando o modelo local de embedding: {model_name}")
        self.model = SentenceTransformer(model_name, trust_remote_code=True)
        self.target_dimensions = 256

    def _get_db_connection(self):
        conn = psycopg2.connect(self.db_url)
        register_vector(conn)
        return conn
    
    def _generate_truncated_vector(self, text: str) -> list:
        # Vetor nulo para dados faltantes para que o postgres nao crashe
        if not text or text == "N/A":
            return [0.0] * self.target_dimensions
        
        prompted_text = f"Represent this instagram profile characteristic for search: {text}"

        raw_embedding = self.model.encode(prompted_text)

        truncated_embedding = raw_embedding[:self.target_dimensions].tolist()
        return truncated_embedding
    
    def ingest_profile(self, username: str, profile_url: str, metadata: dict, semantics_json: str):
        # Basicamente parsear o output do Gemma 4 e gerar multi-vetores e inserir tudo no Postgres
        try:
            print("a")
            # Parsear o output da LLM para um dicionario python
            semantic_data = json.loads(semantics_json).get("semantics", {})

            # Construir os chunks de texto para os indexes dos vetores
            niche_text = json.dumps(semantic_data.get("core_niche", {})) + " " + \
                         json.dumps(semantic_data.get("audience_inference", {}))
            
            appearance_text = json.dumps(semantic_data.get("physical_appearance", {}))

            aesthetic_text = json.dumps(semantic_data.get("visual_aesthetic", {})) + " " + \
                             json.dumps(semantic_data.get("brand_safety", {}))
            
            # Gerar os vetores de 256 dimensoes
            vector_niche = self._generate_truncated_vector(niche_text)
            vector_appearance = self._generate_truncated_vector(appearance_text)
            vector_aesthetic = self._generate_truncated_vector(aesthetic_text)

            # Extrair os filtros hard relacionais
            commercial = semantic_data.get("commercial_classification", {})
            brand_safety = semantic_data.get("brand_safety", {})

            # Inserir no Postgres
            conn = self._get_db_connection()
            cursor = conn.cursor()

            insert_query = """
            INSERT INTO instagram_profiles
            (username, profile_url, entity_type, monetization_platforms,
             resolved_outbound_links, vibe_and_controversy, full_semantic_json,
             vector_niche, vector_appearance, vector_aesthetic)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            ON CONFLICT (username) DO UPDATE
            SET full_semantic_json = EXCLUDED.full_semantic_json,
                vector_niche = EXCLUDED.vector_niche,
                vector_appearance = EXCLUDED.vector_appearance,
                vector_aesthetic = EXCLUDED.vector_aesthetic;
            """

            cursor.execute(insert_query, (
                username,
                profile_url,
                commercial.get("entity_type"),
                metadata.get("monetization_platforms", []),
                metadata.get("resolved_outbound_links", []),
                brand_safety.get("vibe_and_controversy"),
                json.dumps(semantic_data),
                vector_niche,
                vector_appearance,
                vector_aesthetic
            ))

            conn.commit()
            cursor.close()
            conn.close()
            print(f"[SUCESSO] Perfil {username} ingerido com sucesso no banco de dados vetorial.")

        except Exception as e:
            print(f"[Erro / embedder.py] Falha ao ingerir perfil {username}: {e}")