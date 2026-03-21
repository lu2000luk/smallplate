#[macro_use]
extern crate rocket;

use rocket::response::Redirect;
use serde::Deserialize;

const LINK_BASE: &str = "https://link.example.com";
const PLATE_ID: &str = "1";

#[derive(Deserialize)]
struct ResolveEnvelope {
    data: ResolveData,
}

#[derive(Deserialize)]
struct ResolveData {
    destination: String,
}

#[get("/u/<id>/<tail..>?<q..>")]
async fn redirect(id: &str, tail: std::path::PathBuf, q: Option<String>) -> Result<Redirect, rocket::http::Status> {
    let mut resolve = format!("{}/{}/resolve/{}", LINK_BASE, PLATE_ID, id);
    if !tail.as_os_str().is_empty() {
        resolve.push('/');
        resolve.push_str(&tail.to_string_lossy());
    }
    if let Some(query) = q {
        if !query.is_empty() {
            resolve.push('?');
            resolve.push_str(&query);
        }
    }

    let response = reqwest::get(resolve)
        .await
        .map_err(|_| rocket::http::Status::BadGateway)?;
    if !response.status().is_success() {
        return Err(rocket::http::Status::NotFound);
    }
    let payload: ResolveEnvelope = response
        .json()
        .await
        .map_err(|_| rocket::http::Status::BadGateway)?;
    if payload.data.destination.trim().is_empty() {
        return Err(rocket::http::Status::BadGateway);
    }

    Ok(Redirect::temporary(payload.data.destination))
}

#[launch]
fn rocket() -> _ {
    rocket::build().mount("/", routes![redirect])
}
