package repository

const (
	GetUserQuery = `select id, email, password, name, gender, prefer, fromage, toage, date, 
	case when date <> '' then date_part('year', age(date::date)) else 0 end as age,
	description, imgs from profile where email = $1;`

	GetUserByIdAQuery = `select id, email, password, name, gender, prefer, fromage, toage, date, 
	case when date <> '' then date_part('year', age(date::date)) else 0 end as age,
	description, imgs from profile where id = $1;`

	CreateUserQuery = "INSERT into profile(email,password) VALUES($1,$2) RETURNING id, email, password;"

	UpdateUserQuery = `update profile set name=$2, gender=$3, prefer=$4, fromage=$5, toage=$6, date=$7, description=$8, imgs=$9 where email=$1
RETURNING id, email, password, name, gender, prefer, fromage, toage, date, 
case when date <> '' then date_part('year', age(date::date)) else 0 end as age, description, imgs;`

	DeleteTagsQuery = "delete from profile_tag where profile_id=$1 returning id;"

	GetTagsQuery = "select tagname from tag;"

	GetTagsByIdQuery = `select
							tagname
						from
							profile p
							join profile_tag pt on(pt.profile_id = p.id)
							join tag t on(pt.tag_id = t.id)
						where
							p.id = $1;`

	GetImgsByIDQuery = "SELECT imgs FROM profile WHERE id=$1;"

	InsertTagsQueryFirstPart = "insert into profile_tag(profile_id, tag_id) values"
	InsertTagsQueryParts     = "($1, (select id from tag where tagname=$%d))"

	UpdateImgsQuery = "update profile set imgs=$2 where id=$1 returning id;"

	AddReactionQuery = "insert into reactions(id1, id2, type) values ($1,$2,$3) returning id;"

	GetNextUserForSwipeQuery1 = `select 
									op.id,
									op.email,
									op.password,
									op.name,
									op.date,
									case when date <> '' then date_part('year', age(date::timestamp)) else 0 end as age,
									op.description
								from profile op
								where op.id not in (
									select r.id2
									from reactions r
									where r.id1 = $1
								) and op.id not in (
									select m.id2
									from matches m
									where m.id1 = $1
								) and op.id <> $1
									and op.name <> ''
									and op.date <> ''
									and (case when date <> '' then date_part('year', age(date::timestamp)) else 0 end)>=$2
    								and (case when date <> '' then date_part('year', age(date::timestamp)) else 0 end)<=$3
									`

	GetNextUserForSwipeQueryPrefer = "and op.gender=$4\n"

	Limit = " limit 5;"

	GetUsersForMatchesQuery = `select
									op.id,
									op.email,
									op.password,
									op.name,
									op.date,
									case when op.date <> '' then date_part('year', age(op.date::timestamp)) else 0 end as age,
									op.description
								from profile p
								join matches m on (p.id = m.id1)
								join matches om on (om.id1 = m.id2 and om.id2 = m.id1)
								join profile op on (op.id = om.id1)
								where p.id = $1;`

	GetUsersForMatchesWithSearchingQuery = `select
												op.id,
												op.name,
												op.email,
												op.date,
												case when op.date <> '' then date_part('year', age(op.date::timestamp)) else 0 end as age,
												op.description
											from profile p
											join matches m on (p.id = m.id1)
											join matches om on (om.id1 = m.id2 and om.id2 = m.id1)
											join profile op on (op.id = om.id1)
											where p.id = $1 and LOWER(op.name) like LOWER($2);`

	GetLikesQuery = "select r.id1 from reactions r where r.id2 = $1 and r.type = 1;"

	DeleteLikeQuery = "delete from reactions r where ((r.id1=$1 and r.id2=$2) or (r.id1=$2 and r.id2=$1)) returning id;"

	AddMatchQuery = "insert into matches(id1, id2) values ($1,$2),($2,$1) returning id;"

	GetUserLikes = `select p.id,
						   p.name,
						   p.email,
						   p.date,
						   case when p.date <> '' then date_part('year', age(p.date::timestamp)) else 0 end as age,
						   p.description
					from profile p
					join reactions r on (r.id1 = p.id
										 and r.id2 = $1
										 and r.type = 1
										 and p.name <> ''
										 and p.date <> '');`

	GetMessages = `select message_id, from_id, to_id, text, date
    from message
    where
      ((from_id = $1 and to_id = $2) or (from_id = $2 and to_id = $1)) and message_id < $3
    order by date
    limit 100;`

	GetLastMessage = `select message_id, from_id, to_id, text, date
    from message
    where
      (from_id = $1 and to_id = $2)
      or (from_id = $2 and to_id = $1 )
    order by date desc
    limit 1;`

	SendNessage = `
	insert into message(from_id, to_id, text) values ($1,$2,$3) returning message_id;
	`

	InitChat = `
	insert into message(from_id, to_id) values ($1,$2);
	insert into message(from_id, to_id) values ($2,$1);
	`

	GetChats = `
	select
		p.id as FromUserID, p.name, p.imgs[1] as img
	from
		profile p
		join message m on p.id = m.from_id
		join profile op on op.id = m.to_id
	where (m.from_id=$1 or m.to_id=$1) and (p.id<>$1)
	group by p.id, p.name, p.imgs[1];
	`
)
