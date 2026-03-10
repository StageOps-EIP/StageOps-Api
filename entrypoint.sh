#!/bin/sh
set -e

COUCH="${COUCHDB_URL}"
DB="${COUCHDB_DB}"
AUTH="${COUCHDB_USER}:${COUCHDB_PASSWORD}"

echo "⏳ Attente de CouchDB..."
until curl -sf -u "${AUTH}" "${COUCH}/_up" > /dev/null 2>&1; do
  sleep 2
done
echo "✅ CouchDB est prêt."

# Créer la base de données si elle n'existe pas
HTTP=$(curl -s -o /dev/null -w "%{http_code}" -u "${AUTH}" "${COUCH}/${DB}")
if [ "$HTTP" = "404" ]; then
  echo "📦 Création de la base '${DB}'..."
  curl -sf -X PUT -u "${AUTH}" "${COUCH}/${DB}"
  echo "✅ Base '${DB}' créée."
else
  echo "✅ Base '${DB}' existe déjà."
fi

# Créer le design document users (vue by_email) si absent
HTTP=$(curl -s -o /dev/null -w "%{http_code}" -u "${AUTH}" "${COUCH}/${DB}/_design/users")
if [ "$HTTP" = "404" ]; then
  echo "🔍 Création des vues CouchDB..."
  curl -sf -X PUT -u "${AUTH}" "${COUCH}/${DB}/_design/users" \
    -H "Content-Type: application/json" \
    -d '{
      "views": {
        "by_email": {
          "map": "function(doc) { if (doc.type === \"user\" && doc.email) emit(doc.email, null); }"
        }
      }
    }'
  echo "✅ Vues créées."
else
  echo "✅ Vues existent déjà."
fi

echo "🚀 Démarrage du serveur..."
exec ./server
