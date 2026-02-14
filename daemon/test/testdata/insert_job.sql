INSERT INTO jobs(name, enabled, config_json, created_at, updated_at)
VALUES
(
    "simple_backup", 
    1, 
    '{"name":"simple_backup","log":"/path/to/log_folder","compression":false,"command":["ciao"]}', 
    datetime('now'), 
    datetime('now') 
),
( 
    "full_backup", 
    0, 
    '{"name":"full_backup","log":"/path/to/log_folder1","compression":true,"command":["ls"]}', 
    datetime('now'), 
    datetime('now') 
);