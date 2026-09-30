# YNAFB — application architecture

YNAFB ("You Need A F** Budget") is a single-binary personal budgeting app.

There is a thin frontend built on top of the backend rest api. 
The overall goal of the application is for all calculations of happen on the backend, 
with as little logic to exist on the forntend as possible.

The stack is as follows:

1. SQLite: the database, with the schema defined in the migrations directory and
    CRUD queires defined in the queires directory.
    - Data integraty checks exist here, making sure unrepresentable states are impossible
    - No calculations should exist in this layer
2. sqlc: Typesafe go wrappers for the SQLite queires
3. domain layer: where the logic exists
    - Translating the schema layout into a real world model
    - All complex calculations happen here
    - Rich domain model, utilizes an instance map
    - Utilizes caching per context for most queries
4. http/api: the backend rest API 
    - Generated from api.yml
    - Wrapping the domain layer 
    - Mostly request deserializtion and response serializtion 
5. server: handles auth, serving the api and hosting the static frontend
6. frontend: the user interface
    - A thin client built on top of the backend API
    - As little complexity as possible 
    - no caching
    - Does as little data calculation as possible
    - DOES handle some user flows such as categorization, ect.
