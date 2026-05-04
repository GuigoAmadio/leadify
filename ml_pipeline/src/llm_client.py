# src/llm_client.py
import os
from openai import OpenAI
from dotenv import load_dotenv

# Load environment variables
load_dotenv()

class VisionLLMClient:
    def __init__(self):
        # This seamlessly connects to your local Ollama now, and vLLM later
        self.client = OpenAI(
            base_url=os.getenv("LLM_BASE_URL", "http://localhost:11434/v1"),
            api_key="local-no-key" # Required by SDK, but ignored by local servers
        )
        self.model_name = os.getenv("LLM_MODEL_NAME", "gemma4")

    def extract_semantics(self, profile_text: str, base64_image: str) -> str:
        """
        Passes the scraped text and the 3x3 image collage to Gemma 4
        to extract the nuanced semantic JSON.
        """
        
        system_prompt = """
        You are an expert profiling AI. Analyze the provided Instagram profile text and the collage of the user's latest images.
        Extract the semantic details into the EXACT JSON structure below. 
        If the account does not feature a specific human consistently, set all fields inside 'physical_appearance' to 'N/A'.
        
        {
        "semantics": {
            "core_niche": {
            "primary_industry": "Broad category (e.g., 'Fitness', 'B2B SaaS', 'Finance', 'Beauty').",
            "micro_niche": "Highly specific focus (e.g., 'Kettlebell mobility for seniors', 'Cold email deliverability')."
            },
            "commercial_classification": {
            "entity_type": "Classify as: 'Personal Brand/Creator', 'Faceless Theme Page', 'E-commerce Brand', 'Local Brick & Mortar', or 'B2B Agency'.",
            "monetization_model": "How they likely make money (e.g., 'Sponsored Posts', 'Digital Courses/Info Products', 'Physical Products', 'Lead Gen/Consulting').",
            "call_to_action": "The primary action they want users to take based on their bio (e.g., 'Subscribe to Newsletter', 'Buy my Course', 'Book a Call')."
            },
            "audience_inference": {
            "target_demographic": "Who this content is made for (e.g., 'Male tech professionals 20-35', 'Stay-at-home mothers').",
            "target_income_level": "Perceived audience budget based on aesthetics and products (e.g., 'Premium/Luxury', 'Mid-market', 'Budget/Free-seekers')."
            },
            "brand_safety": {
            "vibe_and_controversy": "Is the content 'Brand Safe/Corporate', 'Edgy/Polarizing', or 'Explicit/Not Safe For Work'?",
            "production_value": "Rate the visual quality: 'High-end Studio', 'Prosumer/Good Lighting', or 'Amateur/Raw'."
            },
            "physical_appearance": {
            "demographics": {
                "gender_presentation": "e.g., Female, Male, Androgynous.",
                "perceived_age_range": "e.g., 18-24, 25-30, 30s.",
                "perceived_ethnicity": "e.g., Latina, Caucasian, East Asian. Output 'Unknown' if ambiguous."
            },
            "body_architecture": {
                "build_and_type": "e.g., Petite, curvy, athletic, muscular, slim, plus-size.",
                "facial_features": "e.g., Sharp jawline, freckles, full lips. Mention hair color and style here (e.g., Long blonde hair, short brunette bob)."
            },
            "modifications": {
                "tattoos": "Describe placement and style (e.g., 'Heavy traditional sleeve on left arm, chest piece') or 'None visible'.",
                "piercings": "Describe placement (e.g., 'Septum, multiple ear piercings') or 'None visible'."
            }
            },
            "visual_aesthetic": {
            "color_palette": "Dominant colors and lighting.",
            "environment": "Common background settings (e.g., 'Commercial gym', 'Home office', 'Outdoor urban')."
            },
            "nationality_language": {
                "profile_language": "Language used in bio and captions (e.g., English, Spanish, Portuguese).",
                "country_of_origin": "If explicityly stated or strongly implied (e.g., 'Based in Brazil', 'NYC-based'). Output 'Unknown' if not clear."
            }
        }
        }
        """

        try:
            response = self.client.chat.completions.create(
                model=self.model_name,
                messages=[
                    {"role": "system", "content": system_prompt},
                    {
                        "role": "user",
                        "content": [
                            {"type": "text", "text": f"Profile Text Data:\n{profile_text}\n\nAnalyze this data and the attached image grid."},
                            {
                                "type": "image_url",
                                "image_url": {
                                    # Formats the base64 string for the vision model
                                    "url": f"data:image/jpeg;base64,{base64_image}"
                                }
                            }
                        ]
                    }
                ],
                # Force JSON output mode
                response_format={"type": "json_object"},
                temperature=0.2 # Keep hallucinations low
            )
            
            return response.choices[0].message.content
            
        except Exception as e:
            print(f"[ERROR] LLM Inference Failed: {e}")
            return "{}"