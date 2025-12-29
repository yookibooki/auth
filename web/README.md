- **base.html** — shared layout and styling.
- **auth.html** — contains a “type your email” input with a submit button.  
  If the email exists → log in → "type your password" and submit.
  If it does not exist → "We send you a confirmation link." (sign up) 
- **link-sent.html** — displays “We sent you an email.”
- **success.html** — displays “Success! You can close this window.”  
  Used for all confirmation flows.
- **account.html** — minimalist page with three functions:  
  change email, change password, delete account.


**Server-side Contract**

```go
type PageData struct {
  Step string // "email" | "password" | "signup"

  Email string
  Error string
  Message string

  PostEmailURL    string
  PostPasswordURL string

  ChangeEmailURL    string
  ChangePasswordURL string
  DeleteAccountURL  string
}
```

**Page (HTML) Endpoints**

`/auth` -  Auth entry (email step)          
`/link-sent` - “We sent you an email” page       
`/success` - Generic success page              
`/account` - Account page (auth required)      

**Authentication Flow**

`/auth/email`  - Submit email → decide login vs signup
`/auth/login`  - Password login                        
`/auth/signup` - Create signup + send confirmation link
`/confirm`     - Email confirmation callback (token)   
`/logout`      - End session
