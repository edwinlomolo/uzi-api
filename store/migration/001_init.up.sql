CREATE TABLE IF NOT EXISTS routes (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  distance VARCHAR(225) NOT NULL,
  polyline GEOGRAPHY NOT NULL,
  eta TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  first_name VARCHAR(50) NOT NULL,
  last_name VARCHAR(50) NOT NULL,
  phone VARCHAR(20) UNIQUE NOT NULL,
  onboarding BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS products (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name VARCHAR(100) NOT NULL,
  description TEXT NOT NULL,
  weight_class INTEGER NOT NULL,
  icon TEXT NOT NULL,
  relevance INTEGER NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO products (
  name, description, weight_class, icon, relevance
) VALUES
  ('UziX', 'Faster|Cheaper|Eco-friendly', 638, 'https://uzi-images.s3.eu-west-2.amazonaws.com/icons8-bike-50.png', 1),
  ('UziBoda', 'Convenient|On-demand', 1472, 'https://uzi-images.s3.eu-west-2.amazonaws.com/icons8-motorbike-50.png', 2),
  ('Uzito', 'Loading-truck|Medium-sized', 2000, 'https://uzi-images.s3.eu-west-2.amazonaws.com/icons8-truck-50.png', 3);

CREATE TABLE IF NOT EXISTS couriers (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  verified BOOLEAN DEFAULT false,
  status VARCHAR(10) NOT NULL DEFAULT 'ONBOARDING',
  location GEOGRAPHY,
  ratings INTEGER NOT NULL DEFAULT 0,
  points INTEGER NOT NULL DEFAULT 0,
  user_id UUID REFERENCES users ON DELETE CASCADE,
  product_id UUID REFERENCES products ON DELETE CASCADE,
  trip_id UUID,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS couriers_gix ON couriers USING GIST(location);

CREATE TABLE IF NOT EXISTS trips (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  start_location GEOGRAPHY NOT NULL,
  end_location GEOGRAPHY NOT NULL,
  courier_id UUID REFERENCES couriers ON DELETE CASCADE,
  user_id UUID REFERENCES users ON DELETE CASCADE,
  route_id UUID REFERENCES routes ON DELETE CASCADE,
  cost MONEY,
  status VARCHAR(25) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS vehicles (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  mass FLOAT NOT NULL,
  product_id UUID NOT NULL REFERENCES products ON DELETE CASCADE,
  courier_id UUID REFERENCES couriers ON DELETE CASCADE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS uploads (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  type VARCHAR(3) NOT NULL,
  uri TEXT NOT NULL,
  verification VARCHAR(20) NOT NULL DEFAULT 'ONBOARDING',
  courier_id UUID REFERENCES couriers ON DELETE CASCADE,
  user_id UUID REFERENCES users ON DELETE CASCADE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
  id UUID PRIMARY KEY,
  ip VARCHAR NOT NULL,
  user_agent VARCHAR NOT NULL,
  phone VARCHAR(20) UNIQUE NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
