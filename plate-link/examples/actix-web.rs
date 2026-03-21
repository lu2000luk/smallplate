use actix_web::{get, web, App, HttpRequest, HttpResponse, HttpServer, Responder};
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

#[get("/u/{id}/{tail:.*}")]
async fn redirect(path: web::Path<(String, String)>, req: HttpRequest) -> impl Responder {
    let (id, tail) = path.into_inner();
    let query = req.query_string();
    let mut resolve = format!("{}/{}/resolve/{}", LINK_BASE, PLATE_ID, id);
    if !tail.is_empty() {
        resolve.push('/');
        resolve.push_str(&tail);
    }
    if !query.is_empty() {
        resolve.push('?');
        resolve.push_str(query);
    }

    let response = match reqwest::get(resolve).await {
        Ok(r) => r,
        Err(_) => return HttpResponse::BadGateway().body("resolve failed"),
    };

    if !response.status().is_success() {
        return HttpResponse::build(response.status()).body("link not found");
    }

    let payload: ResolveEnvelope = match response.json().await {
        Ok(v) => v,
        Err(_) => return HttpResponse::BadGateway().body("invalid resolve payload"),
    };

    if payload.data.destination.trim().is_empty() {
        return HttpResponse::BadGateway().body("missing destination");
    }

    HttpResponse::TemporaryRedirect()
        .append_header(("Location", payload.data.destination))
        .finish()
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    HttpServer::new(|| App::new().service(redirect))
        .bind(("127.0.0.1", 8080))?
        .run()
        .await
}
