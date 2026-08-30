import { Pool } from "pg";

import {
  postgresApplicationName,
  postgresConfig,
  postgresIdleTimeoutMilliseconds,
  postgresPoolMaximumSize,
} from "../../configs/postgres.js";

export const postgresPool = new Pool({
  host: postgresConfig.host,
  port: postgresConfig.port,
  user: postgresConfig.user,
  password: postgresConfig.password,
  database: postgresConfig.database,
  application_name: postgresApplicationName,
  max: postgresPoolMaximumSize,
  idleTimeoutMillis: postgresIdleTimeoutMilliseconds,
});
