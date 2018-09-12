create database if not exists gopics;
create table if not exists gopics.albums (id integer NOT NULL AUTO_INCREMENT, name varchar(255) NOT NULL default '', dirname varchar(255) NOT NULL, description varchar(255) NOT NULL default '', images_count integer NOT NULL default 0, updated_at datetime NOT NULL, created_at datetime NOT NULL, primary key(id));
create table if not exists gopics.images (id integer NOT NULL AUTO_INCREMENT, album_id integer NOT NULL, filename varchar(255) NOT NULL, maker varchar(255) NOT NULL, model varchar(255) NOT NULL, lens_maker varchar(255) NOT NULL, lens_model varchar(255) NOT NULL, took_at datetime NOT NULL, f_number float NOT NULL, focal_length integer NOT NULL, iso integer NOT NULL, latitude float NOT NULL, longitude float NOT NULL, updated_at datetime NOT NULL, created_at datetime, primary key(id));
