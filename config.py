import os

from dotenv import load_dotenv

load_dotenv()

phone = os.getenv('PHONE')
api_id = os.getenv('API_ID')
api_hash = os.getenv('API_HASH')
source_group_id = os.getenv('SOURCE_GROUP_ID')
destination_group_id = os.getenv('DESTINATION_GROUP_ID')
