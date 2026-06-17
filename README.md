# Welcome To Chirpy

## Project Overview

Chirpy is a RESTful API backend built to emulate a popular social media platform. It's complete with routing endpoints for several HTTP methods for creating, reading, updating and deleting things from a Postgresql database. and even includes a webhook endpoint for a simulated payment processor.

### Endpoint Usage

GET `/api/chirps`

Retrieves a list of chirps. Parameters are optional and used for filtering or ordering.

| Parameter | Type   | Description                                                 | Default     | Example                                                        |
| --------- | ------ | ----------------------------------------------------------- | ----------- | -------------------------------------------------------------- |
| author_id | UUID   | Filters chirps to only those created by the specified user. | All authors | GET /api/chirps?author_id=2d7de0b5-1137-49b2-9133-42a36e3f1278 |
| sort      | string | Sets the order of the results. Options: asc, desc.          | asc         | GET /api/chirps?sort=asc&author_id={uuid}                      |
