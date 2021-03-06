/**
 * barista.thrift
 * IDL for Barista
 *
 * @author: Anant Bhardwaj
 * @date: 03/23/2014
 */

namespace cpp barista
namespace go barista
namespace java barista
namespace py barista


/* Barista constants */

// version info
const double VERSION = 0.1


/* Database Connection */

// connection parameters
struct ConnectionParams {
  1: optional string client_id,
  2: optional string seq_id,
  3: optional string user,
  4: optional string password,
  5: optional string database
}

// connection info -- must be passed in every execute_sql call
struct Connection {
  1: optional string client_id,
  2: optional string seq_id,
  3: optional string user,
  4: optional string database
}


/* ResultSet */

// A tuple
struct Tuple {
  1: optional list <binary> cells
}

// A result set (list of tuples)
struct ResultSet {
  1: required bool status,
  2: Connection con,
  3: optional i32 row_count,
  4: optional list <Tuple> tuples,
  5: optional list <string> field_names,
  6: optional list <string> field_types
}


/* Barista Exceptions */

// Database Exception
exception DBException {
  1: optional i32 errorCode,
  2: optional string message,
  3: optional string details
}


/* Barista RPC APIs */

service Barista {
  double get_version()

  Connection open_connection (1: ConnectionParams con_params)
      throws (1: DBException ex)

  ResultSet execute_sql (1: Connection con, 2: string query,
      3: list <binary> query_params) throws (1: DBException ex)

  ResultSet execute_sql_txn (1: Connection con, 2: string query,
      3: list <binary> query_params) throws (1: DBException ex)

  void begin_txn (1: Connection con)
      throws (1: DBException ex)

  void commit_txn (1: Connection con)
      throws (1: DBException ex)

  void rollback_txn (1: Connection con)
      throws (1: DBException ex)

  void close_connection (1: Connection con)
      throws (1: DBException ex)
}
