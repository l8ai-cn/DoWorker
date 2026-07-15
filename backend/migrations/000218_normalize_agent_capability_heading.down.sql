DO $$
BEGIN
  RAISE EXCEPTION
    'Migration 000218 is irreversible; restore from the pre-migration backup instead of running down';
END
$$;
