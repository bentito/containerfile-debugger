# Use an official Python runtime as a parent image
FROM python:3.10-slim-buster

# Set the working directory in the container
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . /app

# Install any needed packages
RUN pip install --no-cache-dir Flask

# Make port 80 available to the world outside this container
EXPOSE 80

# Define environment variable
ENV NAME World

# Add a simple Flask app
RUN echo "from flask import Flask\napp = Flask(__name__)\n@app.route('/')\ndef hello():\n    return 'Hello, ${NAME}!'\n\nif __name__ == '__main__':\n    app.run(host='0.0.0.0', port=80)" > app.py

# Run app.py when the container launches
CMD ["python", "app.py"]

