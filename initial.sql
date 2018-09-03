create database if not exists gopics;
create table if not exists gopics.albums (id integer NOT NULL AUTO_INCREMENT, name varchar(255), dirname varchar(255), images_count integer, updated_at datetime, created_at datetime, primary key(id));
create table if not exists gopics.images (id integer NOT NULL AUTO_INCREMENT, album_id integer, filename varchar(255), model varchar(255), lens varchar(255), took_at datetime, f_number float, focal_length integer, iso integer, latitude float, longitude float, primary key(id));
