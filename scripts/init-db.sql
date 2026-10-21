-- Create application schema
CREATE SCHEMA IF NOT EXISTS eng;

-- Create metabase database for Metabase storage
SELECT 'CREATE DATABASE metabase'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'metabase')\gexec
