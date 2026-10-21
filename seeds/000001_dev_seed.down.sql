-- Remove all data seeded by 000003_dev_seed.up.sql
-- Order: child tables first to avoid FK violations

DELETE FROM deadline_history WHERE id IN (1,2,3,4,5,6,7);
DELETE FROM status_history   WHERE id IN (1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25);
DELETE FROM comments         WHERE id IN (1,2,3,4,5,6,7,8,9,10,11,12,13,14,15);
DELETE FROM risks            WHERE id BETWEEN 1 AND 105;
DELETE FROM preferences      WHERE person_id IN (1,5,10,13,29);
DELETE FROM persons          WHERE person_id IN (1,5,10,13,15,18,20,22,25,27,29,30,31,32,33);
