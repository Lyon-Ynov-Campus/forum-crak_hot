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
- **JavaScript** pour une petite partie du front
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
 

**ATTENTION LORS DES CORRECTIONS ET DU DÉPLOIEMENT :**

- **Gestion des messages de confirmation ("Success") :** Lors de l'envoi de certaines requêtes (comme la publication d'un nouveau commentaire), le serveur renvoie un message brut de succès au format JSON indiquant que l'action a été correctement enregistrée en base de données
- **Procédure obligatoire de rafraîchissement :** Dès que ce message de confirmation s'affiche sur fond noir, le correcteur ou l'utilisateur doit **obligatoirement effectuer un retour en arrière avec la flèche de navigation du navigateur, puis exécuter un `Ctrl + R` (ou `Ctrl + F5`)** pour actualiser le cache. Cette manipulation permet de recharger proprement le forum et de voir la publication (ou le commentaire) s'afficher directement à l'écran
- **Réinitialisation de session (Debug) :** Si vous souhaitez forcer la déconnexion ou tester le comportement d'un nouvel utilisateur sans passer par le bouton de déconnexion, ouvrez l'inspecteur du navigateur (F12), allez dans l'onglet *Application/Stockage*, puis supprimez manuellement le cookie `session_token`. Vous pouvez ensuite nettoyer la base de données en supprimant le fichier SQLite dans VSCode (il se recréera automatiquement à blanc au prochain `go run .`)

## Guide d’utilisation du forum

### Connexion / Inscription
- Un système d'authentification complet est mis en place avec pseudo et email unqiue, tout en respectant les recommandations de la CNIL et la RGPD.
- Possibilité de gérer les informations de son compte.
- Les actions comme poster, commenter ou liker sont disponible qu'une fois connecté.
- Le tout avec une gestion d'erreur complète

- **NOTE IMPORTANTE** : Par manque de temps, nous n'avons pas pû terminer le système de "mot de passe oublié" dans son ensemble, autrement dit aucun mail est envoyé à l'utilisateur. En revanche, en attendant, le lien était envoyé en terminal et tout le système lié était fonctionnel (comme présenté lors de la soutenance), or ici, comme le système est déployé, le lien est envoyé coté console azure ce qui rend le lien de réinitialisation du mot de passe innacessible bien que fonctionnel.

### Recherche
- Vous pouvez rechercher un post par titre via la barre de recherche.  
La recherche fonctionne avec le tri et le filtre (date, likes, commentaires).
- Vous pouvez rechercher un membre a l'aide de son pseudo dans la partie "Réseau".

### Espace Réseau (Fonctionnalité en cours d'unification)

L'Espace Réseau a pour but de centraliser l'annuaire des membres de la communauté Crack'HOT

 - **Ce que l'on peut faire :** Une barre de recherche asynchrone (API) permet de taper le pseudo d'un membre. Si le membre existe, une carte de profil générée dynamiquement en JavaScript apparaît à l'écran, affichant son pseudonyme et sa photo de profil (ou un avatar par défaut)
 - **Ce qui ne fonctionne pas encore / Limitations :** Le bouton "Voir le profil" souffre actuellement d'un problème technique lors du transfert de l'identifiant numérique (`id`) entre le JavaScript et le serveur Go, générant parfois un ID égal à `0`. Par conséquent, cliquer sur ce bouton provoque une erreur 404 (page introuvable) ou redirige l'utilisateur en boucle vers la page d'accueil réseau. La page finale (`seeUser.html`), qui est censée lister l'historique complet des posts et des commentaires d'un autre utilisateur, est prête mais reste bloquée par ce bug de transmission

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
forum-crackhot-gaevdhate6bha8b8.francecentral-01.azurewebsites.net
