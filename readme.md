<!-- Improved compatibility of back to top link -->
<a id="readme-top"></a>

<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/bensabler/milos-residence">
    <img src="static/images/hero-milo.png" alt="Milo’s Residence" width="120" height="120">
  </a>

  <h3 align="center">Milo’s Residence</h3>

  <p align="center">
    A small Go web app demonstrating routing, templates, sessions, CSRF, and basic form handling.
    <br />
    <a href="#about-the-project"><strong>Learn more »</strong></a>
  </p>
</div>

<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About the Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#testing">Testing</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#acknowledgments">Acknowledgments</a></li>
  </ol>
</details>

## About the Project

Milo’s Residence is a focused, server-rendered Go application. It showcases:

- HTTP routing with middleware
- Server-side HTML templates with a cache toggle for dev/prod
- Session-based flash/warn/error messaging
- CSRF protection for POST forms
- Form validation helpers and table-driven tests

Feature pages:
- Home, About, Photos, Contact
- “Snooze spots”: Golden Haybeam Loft, Window Perch Theater, Laundry Basket Nook
- Reservations: `/make-reservation` (GET/POST), `/reservation-summary`
- Availability: `/search-availability` (GET/POST), `/search-availability-json` (POST JSON)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Built With

- Go 1.24.6
- Router: [chi]
- Sessions: [scs]
- CSRF: [nosurf]
- Templates: Go `html/template`

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Getting Started

### Prerequisites
- Go 1.24.6 (or compatible with your `go.mod`)
