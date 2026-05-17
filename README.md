[![Review Assignment Due Date](https://classroom.github.com/assets/deadline-readme-button-22041afd0340ce965d47ae6ef1cefeee28c7c493a6346c4f15d667ab976d596c.svg)](https://classroom.github.com/a/iWb6fRWH)

# **Crack'HOT**

## Présentation du projet

Crack'HOT est un forum dédiée aux passionnés de cuisine épicée.  
Elle permet aux utilisateurs de :

- créer un compte et se connecter
- publier des posts (recettes, adresses, conseils…)
- commenter les publications
- liker les posts
- rechercher des sujets
- rechercher des membres
- trier et filtrer les résultats (date, likes, commentaires)
- Modification ou suppression des données

Le projet repose sur :

- **Go (Golang)** pour le backend
- **HTML/CSS** pour le front
- **SQLite** pour la base de données
- **API** pour la gestion des posts, commentaires et utilisateurs

## Tuto d'installation
Pour tester ce projet, suivez les étapes suivantes :

- Ouvrez VSCode et sélectionnez le dossier dans lequel vous souhaitez installer les fichiers
- Ouvrez le terminal intégré et faite la commande "git clone [lien github du projet]"
- Après l'installation des fichiers, faites la commande "go mod tidy" au même endroit, pour vérifier et compléter les dépendances nécessaires
- Puis la commande "go get github.com/mattn/go-sqlite3 " afin d'installer la bibliothèque pour la base de donnée
- Enfin, vous n'avez plus qu'à faire la commande "go run ." et vous rendre sur votre navigateur sur l'adresse "http://localhost:8080"

## Pour une meilleure utilisation
- Ne pas revenir en arrière avec les flèches de naviguations mais plutôt a l'aide des boutons(cela peut fausser les résultats ou le chargement de la page)
- Si vous souhaiter quitter la page sans vous déconnecter, aller dans inspection, puis supprimez manuellement le user. A la suite de cela, effacer la base de donnée dans VSCode (elle se rechargera automatiquement avec le go run .)

## Guide d’utilisation du forum

### Recherche
- Vous pouvez rechercher un post par titre via la barre de recherche.  
La recherche fonctionne avec le tri et le filtre (date, likes, commentaires).
- Vous pouvez rechercher un membre a l'aide de son pseudo dans la partie "Réseau".

### Création de post
Accessible uniquement si vous êtes connecté.

### Commentaires
Chaque post peut être commenté, et les commentaires sont comptabilisés automatiquement.

### Likes
Chaque post peut être liké une seule fois par utilisateur et sont comptabilisés automatiquement comme les commentaires.

### Profil
L'utilisateur peut modifier ou supprimer ses données et ou ses posts et commentaires dns la partie profil. Cela s'applique directement et automatiquement sur la base de données.

## Utilisation d'IA respectueuse concernant Marjane :
- Les lignes faites par IA sont commenter directement dans le code.

## Déploiement lien :
https://forum-crackhot-gaevdhate6bha8b8.francecentral-01.azurewebsites.net/forum
