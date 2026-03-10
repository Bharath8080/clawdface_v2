import os
import json
import glob
from supabase import create_client, Client
from dotenv import load_dotenv

load_dotenv(os.path.join(os.getcwd(), ".env"))

SUPABASE_URL = os.getenv("NEXT_PUBLIC_SUPABASE_URL")
SUPABASE_KEY = os.getenv("SUPABASE_SERVICE_ROLE_KEY") or os.getenv("NEXT_PUBLIC_SUPABASE_ANON_KEY")

print(f"DEBUG: URL={SUPABASE_URL}")
print(f"DEBUG: KEY={SUPABASE_KEY}")

if not SUPABASE_URL or not SUPABASE_KEY:
    print("Error: NEXT_PUBLIC_SUPABASE_URL or SUPABASE_KEY (SERVICE_ROLE or ANON) not found in .env")
    exit(1)

supabase: Client = create_client(SUPABASE_URL, SUPABASE_KEY)

def migrate_verified_users():
    users_file = r"c:\Users\homeu\clawdface\frontend\data\verified-users.json"
    if os.path.exists(users_file):
        with open(users_file, "r") as f:
            users = json.load(f)
            for email in users:
                print(f"Migrating verified user: {email}")
                supabase.from_("verified_users").upsert({"email": email.lower()}).execute()

def migrate_user_configs():
    config_dir = r"c:\Users\homeu\clawdface\frontend\data\user-configs"
    if os.path.exists(config_dir):
        for file_path in glob.glob(os.path.join(config_dir, "*.json")):
            email = os.path.basename(file_path).replace(".json", "")
            with open(file_path, "r") as f:
                config = json.load(f)
                print(f"Migrating config for: {email}")
                supabase.from_("user_configs").upsert({
                    "email": email.lower(),
                    "openclaw_url": config.get("openclawUrl"),
                    "gateway_token": config.get("gatewayToken"),
                    "session_key": config.get("sessionKey")
                }).execute()

if __name__ == "__main__":
    print("Starting migration to Supabase...")
    migrate_verified_users()
    migrate_user_configs()
    print("Migration complete!")
