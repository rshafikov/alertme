package migrations

const CreateMetricsType = `
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'metrics_type') THEN
        CREATE TYPE metrics_type AS ENUM ('gauge', 'counter');
    END IF;
END $$;
`
const CreateMetricsTable = `
	CREATE TABLE IF NOT EXISTS metrics (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		value DOUBLE PRECISION,
		delta BIGINT,
		type TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
`
