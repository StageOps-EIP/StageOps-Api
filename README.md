# StageOps Backend

API centrale de la plateforme StageOps.

Ce service constitue le cœur logique du système et gère l’ensemble des données, des règles métiers et des services d’intelligence technique.

## Responsabilités

- Gestion des utilisateurs et rôles techniques
- Gestion des projets scéniques
- Inventaire matériel
- Suivi d’état opérationnel
- Gestion des incidents
- Synchronisation multi-clients
- Calcul de l’usure résiduelle des lampes (RUL)

## Architecture

Architecture modulaire orientée domaine.

src/
  modules/
    auth/
    users/
    projects/
    equipment/
    lighting/
    sound/
    incidents/
    rul-engine/
  database/
  middleware/
  config/

## Algorithme RUL (Remaining Useful Life)

Le module RUL estime la durée de vie restante des sources lumineuses en fonction de :

- heures d’utilisation
- cycles d’allumage
- température de fonctionnement
- historique de maintenance

Objectif : maintenance prédictive du matériel scénique.

## Stack technique

- Node.js
- API REST / GraphQL
- PostgreSQL ou CouchDB
- JWT Authentication
- Validation middleware

## Installation

### Prérequis
Node.js >= 18  
Base de données PostgreSQL ou CouchDB  

### Setup

git clone <repo>
cd stageops-backend
npm install

Créer `.env`

DATABASE_URL=
JWT_SECRET=
PORT=3000

### Lancement

npm run dev

## Sécurité

- Authentification tokenisée
- Permissions par rôle technique
- Validation des entrées

## Tests

npm run test

## Objectif MVP

Fournir une API stable pour la gestion des opérations techniques scéniques.

## Licence

Projet académique.
