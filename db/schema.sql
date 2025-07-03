PRAGMA foreign_keys = ON;

-- an album of various posts, can be across contexts
create table if not exists collections(
  id uuid primary key,
  title text not null,
  slug text not null,
  description text not null,
  collection_description_post_id uuid,
  clicked_count integer default 0,
  created datetime default current_timestamp,
  foreign key (collection_description_post_id) references posts(id) on delete set null
);

-- contexts (like krug is usually tech stuff, draw_dngn is usually art stuff)
create table if not exists contexts (
  id uuid primary key,
  title text not null,
  description text not null,
  context_description_post_id uuid,
  created datetime default current_timestamp,
  foreign key (context_description_post_id) references posts(id) on delete set null
);

-- post types are like video, blog, gallery
create table if not exists posts (
  id uuid primary key,
  type text,
  visible boolean default false,
  title text not null,
  subtitle text,
  content text not null,
  slug text unique not null,
  permalink text unique not null,
  created datetime default current_timestamp
);

create table if not exists tags(
  id uuid primary key,
  title text unique not null,
  searched_count integer default 0,
  created datetime default current_timestamp
);

create table if not exists post_tags(
  id uuid primary key,
  post_id uuid not null,
  tag_id uuid not null,
  created datetime default current_timestamp,
  foreign key (post_id) references posts(id) on delete cascade,
  foreign key (tag_id) references tags(id) on delete cascade,
  unique(post_id, tag_id)
);

create table if not exists context_posts(
  id uuid primary key,
  post_id uuid not null,
  context_id uuid not null,
  created datetime default current_timestamp,
  foreign key (post_id) references posts(id) on delete cascade,
  foreign key (context_id) references contexts(id) on delete cascade,
  unique(post_id, context_id)
);

create table if not exists collection_posts(
  id uuid primary key,
  collection_id uuid not null,
  post_id uuid not null,
  "order" integer default 0,
  created datetime default current_timestamp,
  foreign key (collection_id) references collections(id) on delete cascade,
  foreign key (post_id) references posts(id) on delete cascade,
  unique(collection_id, post_id)
);

-- handles the linking to social accounts
-- for the cron that deploys the post to the account
create table if not exists crosspost_jobs (
  id uuid primary key,
  platform text not null,
  post_id uuid not null,
  created datetime default current_timestamp,
  foreign key (post_id) references posts(id) on delete cascade
);

create table if not exists instagram_details(
  id uuid primary key,
  account_name text not null unique,
  access_key text not null,
  account_id text not null unique,
  has_threads_link boolean not null default false,
  created datetime default current_timestamp
);

create table if not exists instagram_posts(
  id uuid primary key,
  post_id uuid not null,
  instagram_url text,
  last_synced datetime,
  created datetime default current_timestamp,
  foreign key (post_id) references posts(id) on delete cascade
);

create table if not exists threads_posts(
  id uuid primary key,
  post_id uuid not null,
  last_synced datetime,
  created datetime default current_timestamp,
  foreign key (post_id) references posts(id) on delete cascade
);

-- generic cron maintainer
create table if not exists jobs(
  id uuid primary key,
  job_name text not null,
  status text not null,
  result_message text,
  completed datetime,
  created datetime default current_timestamp
);

create table if not exists links(
  id uuid primary key,
  title text,
  is_local_href boolean default false,
  href text,
  image_url text,
  find_out_more_href text,
  click_count integer default 0,
  is_visible boolean default false,
  "order" integer default 0,
  created datetime default current_timestamp
);

-- collections indices
create index if not exists idx_collections_slug on collections(slug);
create index if not exists idx_collections_clicked_count on collections(clicked_count desc);
create index if not exists idx_collections_created on collections(created desc);

-- contexts indices  
create index if not exists idx_contexts_title on contexts(title);
create index if not exists idx_contexts_created on contexts(created desc);

-- posts indices (most important for performance)
create index if not exists idx_posts_slug on posts(slug);
create index if not exists idx_posts_permalink on posts(permalink);
create index if not exists idx_posts_visible on posts(visible);
create index if not exists idx_posts_type on posts(type);
create index if not exists idx_posts_created on posts(created desc);
create index if not exists idx_posts_visible_created on posts(visible, created desc);
create index if not exists idx_posts_type_visible on posts(type, visible);

-- tags indices
create index if not exists idx_tags_title on tags(title);
create index if not exists idx_tags_searched_count on tags(searched_count desc);

-- junction table indices for joins
create index if not exists idx_post_tags_post_id on post_tags(post_id);
create index if not exists idx_post_tags_tag_id on post_tags(tag_id);

create index if not exists idx_context_posts_post_id on context_posts(post_id);
create index if not exists idx_context_posts_context_id on context_posts(context_id);

create index if not exists idx_collection_posts_collection_id on collection_posts(collection_id);
create index if not exists idx_collection_posts_post_id on collection_posts(post_id);
create index if not exists idx_collection_posts_order on collection_posts(collection_id, "order");

-- crosspost indices
create index if not exists idx_crosspost_jobs_platform on crosspost_jobs(platform);
create index if not exists idx_crosspost_jobs_post_id on crosspost_jobs(post_id);
create index if not exists idx_crosspost_jobs_created on crosspost_jobs(created desc);

-- instagram indices
create index if not exists idx_instagram_details_account_name on instagram_details(account_name);
create index if not exists idx_instagram_posts_post_id on instagram_posts(post_id);
create index if not exists idx_instagram_posts_last_synced on instagram_posts(last_synced);

-- threads indices
create index if not exists idx_threads_posts_post_id on threads_posts(post_id);
create index if not exists idx_threads_posts_last_synced on threads_posts(last_synced);

-- jobs indices
create index if not exists idx_jobs_job_name on jobs(job_name);
create index if not exists idx_jobs_status on jobs(status);
create index if not exists idx_jobs_created on jobs(created desc);
create index if not exists idx_jobs_completed on jobs(completed desc);

-- links indices  
create index if not exists idx_links_is_visible on links(is_visible);
create index if not exists idx_links_order on links("order");
create index if not exists idx_links_visible_order on links(is_visible, "order");
create index if not exists idx_links_click_count on links(click_count desc);
create index if not exists idx_links_is_local_href on links(is_local_href);
