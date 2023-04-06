SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: manveru; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA manveru;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: derivation_outputs; Type: TABLE; Schema: manveru; Owner: -
--

CREATE TABLE manveru.derivation_outputs (
    drv integer NOT NULL,
    id text NOT NULL,
    path text NOT NULL
);


--
-- Name: refs; Type: TABLE; Schema: manveru; Owner: -
--

CREATE TABLE manveru.refs (
    referrer integer NOT NULL,
    reference integer NOT NULL
);


--
-- Name: schema_migrations; Type: TABLE; Schema: manveru; Owner: -
--

CREATE TABLE manveru.schema_migrations (
    version character varying(255) NOT NULL
);


--
-- Name: valid_paths; Type: TABLE; Schema: manveru; Owner: -
--

CREATE TABLE manveru.valid_paths (
    id integer NOT NULL,
    path text NOT NULL,
    hash text NOT NULL,
    registration_time timestamp with time zone,
    deriver text,
    nar_size integer,
    ultimate boolean,
    sigs text[],
    ca text
);


--
-- Name: derivation_outputs derivation_outputs_pkey; Type: CONSTRAINT; Schema: manveru; Owner: -
--

ALTER TABLE ONLY manveru.derivation_outputs
    ADD CONSTRAINT derivation_outputs_pkey PRIMARY KEY (drv, id);


--
-- Name: refs refs_pkey; Type: CONSTRAINT; Schema: manveru; Owner: -
--

ALTER TABLE ONLY manveru.refs
    ADD CONSTRAINT refs_pkey PRIMARY KEY (referrer, reference);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: manveru; Owner: -
--

ALTER TABLE ONLY manveru.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: valid_paths valid_paths_path_key; Type: CONSTRAINT; Schema: manveru; Owner: -
--

ALTER TABLE ONLY manveru.valid_paths
    ADD CONSTRAINT valid_paths_path_key UNIQUE (path);


--
-- Name: valid_paths valid_paths_pkey; Type: CONSTRAINT; Schema: manveru; Owner: -
--

ALTER TABLE ONLY manveru.valid_paths
    ADD CONSTRAINT valid_paths_pkey PRIMARY KEY (id);


--
-- Name: index_derivation_outputs; Type: INDEX; Schema: manveru; Owner: -
--

CREATE INDEX index_derivation_outputs ON manveru.derivation_outputs USING btree (path);


--
-- Name: index_reference; Type: INDEX; Schema: manveru; Owner: -
--

CREATE INDEX index_reference ON manveru.refs USING btree (reference);


--
-- Name: index_referrer; Type: INDEX; Schema: manveru; Owner: -
--

CREATE INDEX index_referrer ON manveru.refs USING btree (referrer);


--
-- Name: derivation_outputs derivation_outputs_drv_fkey; Type: FK CONSTRAINT; Schema: manveru; Owner: -
--

ALTER TABLE ONLY manveru.derivation_outputs
    ADD CONSTRAINT derivation_outputs_drv_fkey FOREIGN KEY (drv) REFERENCES manveru.valid_paths(id) ON DELETE CASCADE;


--
-- Name: refs refs_reference_fkey; Type: FK CONSTRAINT; Schema: manveru; Owner: -
--

ALTER TABLE ONLY manveru.refs
    ADD CONSTRAINT refs_reference_fkey FOREIGN KEY (reference) REFERENCES manveru.valid_paths(id) ON DELETE RESTRICT;


--
-- Name: refs refs_referrer_fkey; Type: FK CONSTRAINT; Schema: manveru; Owner: -
--

ALTER TABLE ONLY manveru.refs
    ADD CONSTRAINT refs_referrer_fkey FOREIGN KEY (referrer) REFERENCES manveru.valid_paths(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--


--
-- Dbmate schema migrations
--

INSERT INTO manveru.schema_migrations (version) VALUES
    ('20221120032825');
