import os

from dotenv import load_dotenv

load_dotenv()

phone = os.getenv('PHONE')
api_id = os.getenv('API_ID')
api_hash = os.getenv('API_HASH')
source_group_id = int(os.getenv('SOURCE_GROUP_ID'))
target_group_id = int(os.getenv('TARGET_GROUP_ID'))
