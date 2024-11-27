begin;

create table migrations(
  num int not null unique,
  title varchar(255) not null unique,
  applied bool
);

end;
