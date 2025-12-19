SELECT
  a.id,
  a.role,
  a.candidate_id,
  a.status,
  a.version,
  c.first_name,
  c.last_name,
  c.email
FROM applications a
JOIN candidates c
  ON c.id = a.candidate_id
WHERE c.id = candidate_id;