-- +goose Up
CREATE TABLE IF NOT EXISTS audio (
	id BIGSERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	author TEXT NOT NULL,
	playlist TEXT
);
CREATE TABLE IF NOT EXISTS audio_tags (
	audio_id BIGINT NOT NULL REFERENCES audio(id) ON DELETE CASCADE,
	tag TEXT NOT NULL,
	UNIQUE (audio_id, tag)
);

-- +goose Down
DROP TABLE audio_tags;
DROP TABLE audio;
