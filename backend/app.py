import os
import time
from flask import Flask, request
from werkzeug.utils import secure_filename

app = Flask(__name__)

ALLOWED_IMAGE_FILE_EXTENSIONS = ['png', 'jpg', 'jpeg']

def is_file_extension_allowed(filename: str, allowed_extensions: list[str]) -> bool:
  has_extension = '.' in filename
  if not has_extension:
    return False

  dot_separated = filename.rsplit('.', 1) 
  return dot_separated[1].lower() in allowed_extensions 

@app.route('/image-classification', methods=['POST'])
def image_classification():
  IMAGE_FIELD_NAME = 'image'

  if IMAGE_FIELD_NAME not in request.files:
    return f'{IMAGE_FIELD_NAME} missing from request', 400
  
  image_file = request.files[IMAGE_FIELD_NAME]
  
  if image_file.filename == '':
    return f'{IMAGE_FIELD_NAME} missing from request', 400
  
  if not is_file_extension_allowed(image_file.filename, ALLOWED_IMAGE_FILE_EXTENSIONS):
    return f'{IMAGE_FIELD_NAME} has invalid file extension. Allowed extensions: {", ".join(ALLOWED_IMAGE_FILE_EXTENSIONS)}', 400

  image_file.save(os.path.join('images', str(time.time()) + secure_filename(image_file.filename)))
  return '', 201

@app.errorhandler(404)
def not_found(_):
  return 'Page not found', 404
