package oracle.oauthproxy;

import com.auth0.json.auth.TokenHolder;
import com.auth0.jwk.*;
import com.auth0.jwt.JWT;
import com.auth0.jwt.JWTVerifier;
import com.auth0.jwt.algorithms.Algorithm;
import com.auth0.jwt.exceptions.JWTVerificationException;
import com.auth0.jwt.interfaces.Claim;
import com.auth0.jwt.interfaces.DecodedJWT;
import com.fasterxml.jackson.core.JsonFactory;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import okhttp3.*;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import javax.ws.rs.*;
import javax.ws.rs.core.Context;
import javax.ws.rs.core.MediaType;
import javax.ws.rs.core.NewCookie;
import javax.ws.rs.core.UriInfo;
import java.io.IOException;
import java.io.InputStream;
import java.net.MalformedURLException;
import java.net.URI;
import java.net.URL;
import java.net.URLConnection;
import java.security.PublicKey;
import java.security.interfaces.RSAPrivateKey;
import java.security.interfaces.RSAPublicKey;
import java.util.HashMap;
import java.util.Map;

import static javax.ws.rs.core.Cookie.DEFAULT_VERSION;
import static javax.ws.rs.core.NewCookie.DEFAULT_MAX_AGE;

/**
 * Root resource (exposed at "p" path)
 */
@Path("p") public class OpenIDProxyResource {
  private static OpenIdConfiguration config;
  private final String issuer = Main.opts.issuer;
  private final String clientId = Main.opts.clientId;
  private final String clientSecret = Main.opts.clientSecret;
  private final String redirectUri = Main.opts.redirectUri;
  private final String state = "1234";
  private static final String scope = "openid offline_access";
  private static final String IDTOKEN_COOKIE = "idtoken";
  private static final String ACCESS_TOKEN_COOKIE = "accesstoken";
  private static final String TENANT_ID_COOKIE = "tenant";
  public static final String REDIRECT_COOKIE_NAME = "org";
  public static final String AUTHORIZATION_HEADER = "Authorization";

  @Context UriInfo proxyUri;
  @Context javax.ws.rs.core.Request req;

  @HeaderParam("X-Forwarded-Host")
  String forwardedHost;
  @CookieParam(REDIRECT_COOKIE_NAME)
  String initialUrl;

  private ObjectMapper mapper = new ObjectMapper();
  private static final Logger LOG = LoggerFactory.getLogger(OpenIDProxyResource.class);

  /**
   * Initiates OpenID login flow. Basically does a redirect (via browser) to OpenID server requesting auth code.
   *
   * @param sc
   * @param redirect
   * @return
   */
  @Path("login") @GET public javax.ws.rs.core.Response login(
      @QueryParam("scope") @DefaultValue(value = scope) String sc,
      @HeaderParam("X-Auth-Request-Redirect") String redirect,
      @QueryParam("rd") String rd) {
    HttpUrl httpUrl = HttpUrl.parse(getOpenIdConfig().getAuthorizationEndpoint()).newBuilder()
        .addQueryParameter("client_id", clientId)
        .addQueryParameter("redirect_uri", redirectUri)
        .addQueryParameter("response_type", "code")
        .addQueryParameter("scope", sc)
        .addQueryParameter("state", state)
        .build();
    LOG.debug("login - " + httpUrl.toString());

    String r = redirect != null ? redirect : rd;
    LOG.info(String.format("Initial URL: '%s'", r));

    final NewCookie originalUrl = new NewCookie(REDIRECT_COOKIE_NAME, r, "/",
        "", DEFAULT_VERSION, null, DEFAULT_MAX_AGE, null, false, false);
    LOG.info(String.format("orig URL: '%s'", originalUrl));
    return javax.ws.rs.core.Response.status(javax.ws.rs.core.Response.Status.FOUND).location(httpUrl.uri()).
        cookie(originalUrl).build();
  }

  /**
   * This callback endpoint is registered with oauth client on OpenID server. OpenID server responds to the above request
   * with a redirect (via browser) to this endpoint along with auth code.
   * This one performs exchanging auth code with access token (using okhttp client).
   * After successful completion, the id token is stored as cookie for future requests
   * Also it perform redirect to original url which triggered login flow.
   *
   * @param code
   * @param state
   * @param org
   * @return
   */
  @Path("callback") @GET public javax.ws.rs.core.Response callback(@QueryParam("code") String code, @QueryParam("state") String state,
      @CookieParam(REDIRECT_COOKIE_NAME)  @DefaultValue("/bpm/api/4.0/dp-executions") String org) {
    //exchange
    HttpUrl httpUrl = HttpUrl.parse(getOpenIdConfig().getTokenEndpoint());
    OkHttpClient client = new OkHttpClient();
    LOG.debug("callback " + httpUrl.toString());
    RequestBody formBody = new FormBody.Builder()
        .add("client_id", clientId)
        .add("client_secret", clientSecret)
        .add("grant_type", "authorization_code")
        .add("code", code)
        .add("redirect_uri", redirectUri)
        .build();

    //IDCS needs  clientId and secret as basic auth header
    final String basic = Credentials.basic(clientId, clientSecret);
    Request request = new Request.Builder().url(httpUrl).header(AUTHORIZATION_HEADER, basic).post(formBody).build();
    try {
      Response response = client.newCall(request).execute();
      if (!response.isSuccessful()) {
        LOG.error(response.body().string());
        return javax.ws.rs.core.Response.status(response.code()).entity(response.body()).build();
      }
      LOG.debug(response.toString());
      final ResponseBody body = response.body();
      String payload = body.string();
      LOG.debug(payload);
      TokenHolder tokenHolder = mapper.readValue(payload, TokenHolder.class);

      //now token is here, lets set it as cookie
      final URI location = URI.create(org);
      LOG.debug("o the domain: " + location);
      String domain = getCookieDomain();
      String tenantId = getTenantId();
      final NewCookie idTokenCookie = new NewCookie(IDTOKEN_COOKIE, tokenHolder.getIdToken(), "/", "", DEFAULT_VERSION, null,
          DEFAULT_MAX_AGE, null, false, false);

      final NewCookie accessTokenCookie = new NewCookie(ACCESS_TOKEN_COOKIE, tokenHolder.getAccessToken(), "/", "", DEFAULT_VERSION, null,
          DEFAULT_MAX_AGE, null, false, false);

      final NewCookie tenantIdCooke = new NewCookie(TENANT_ID_COOKIE, tenantId, "/", "", DEFAULT_VERSION, null,
          DEFAULT_MAX_AGE, null, false, false);

      //redirect to original request
      LOG.debug("Tokens are attached to the domain: " + domain);
      return javax.ws.rs.core.Response.status(javax.ws.rs.core.Response.Status.FOUND).location(location)
          .cookie(idTokenCookie)
          .cookie(accessTokenCookie)
          .cookie(tenantIdCooke)
          .build();
    } catch (IOException e) {
      LOG.error(e.getMessage(), e);
    }
    return null;
  }

  private String getCookieDomain() {
    String domain = proxyUri.getBaseUri().getHost();
    LOG.debug("Uri Host: " + domain);
    if (typeOf(domain).equals(OpenIdConfiguration.DomainType.FQDN)) {
      domain = domain.substring(domain.indexOf(".") + 1);
    }else{
      domain = null;
    }
    LOG.debug("Cookie domain: " + domain);
    return domain;
  }

  private String getTenantId() {
    String domain = null;
    if(domain == null && initialUrl != null){
      domain = URI.create(initialUrl).getHost();
    }
    if(domain == null)
      domain = proxyUri.getBaseUri().getHost();
    String tenantId = domain.substring(0, domain.indexOf("."));
    LOG.debug("TenantId: " + tenantId);
    return tenantId;
  }

  private OpenIdConfiguration.DomainType typeOf(String domain) {
    if (domain.contains(":") && !domain.contains(".")) {
      return OpenIdConfiguration.DomainType.IPv6;
    }
    if(domain.matches("^.[0-9]{1,3}/..[0-9]{1,3}/..[0-9]{1,3}/..[0-9]{1,3}")){
      return OpenIdConfiguration.DomainType.IPv4;
    }
    if (domain.contains(".")) {
      return OpenIdConfiguration.DomainType.FQDN;
    }
    return OpenIdConfiguration.DomainType.SHORT_NAME;
  }

  /**
   * This is used by nginx auth_request directive to check if a given request is authenticated.
   * accepts jwt token as cookie or as Authorization header
   */
  @Path("/auth")
  @GET @Produces(MediaType.TEXT_PLAIN)
  public javax.ws.rs.core.Response validateToken(
      @HeaderParam(AUTHORIZATION_HEADER) String authzHeader,
      @CookieParam(IDTOKEN_COOKIE) String idToken,
      @CookieParam(ACCESS_TOKEN_COOKIE) String accessToken) {

    //cookies to upon successful validation
    String tenantId = getTenantId();
    final NewCookie tenantIdCooke = new NewCookie(TENANT_ID_COOKIE, tenantId, "/", "", DEFAULT_VERSION, null,
        DEFAULT_MAX_AGE, null, false, false);

    String token = null;
    if (authzHeader != null) {
      final String[] strings = authzHeader.split("//s+");
      token = strings[1];
    } else if (idToken != null) {
      token = idToken;
    }
    if (token == null || token.isEmpty()) {
      LOG.debug("No token found for the uri: " + proxyUri.getAbsolutePath());
      return javax.ws.rs.core.Response.status(javax.ws.rs.core.Response.Status.UNAUTHORIZED).build();
    }

    if (config.getValidationType().equals(OpenIdConfiguration.ValidationType.INTROSPECTION)) {
      if (introspect(accessToken)) {
        return javax.ws.rs.core.Response.ok().cookie(tenantIdCooke).cookie(tenantIdCooke).build();
      }else{
        return javax.ws.rs.core.Response.status(javax.ws.rs.core.Response.Status.UNAUTHORIZED).build();
      }
    }
    Jwk jwk;
    try {
      JwkProvider provider = new UrlJwkProvider(new URL(getOpenIdConfig().getJwksUri()));

      DecodedJWT jwt = JWT.decode(token);
      final Claim kid = jwt.getHeaderClaim("kid");
      jwk = provider.get(kid.asString());
      final PublicKey publicKey = jwk.getPublicKey();

      RSAPrivateKey privateKey = null;//Get the key instance
      try {
        Algorithm algorithm = Algorithm.RSA256((RSAPublicKey) publicKey, privateKey);

        JWTVerifier verifier = JWT.require(algorithm).withIssuer(issuer).build(); //Reusable verifier instance
        verifier.verify(token);
      } catch (JWTVerificationException e) {
        LOG.error(e.getMessage(), e);
      }

      return javax.ws.rs.core.Response.ok().cookie(tenantIdCooke).build();
    } catch (JwkException e) {
      //IDCS does not expose Jwks endpoint publicly (driven by settings), hence token cannot be evaluated locally
      //perform introspection call
      LOG.error(e.getMessage(), e);
      token = accessToken;
      //try introspection
      if (introspect(token)) {
        config.setValidationType(OpenIdConfiguration.ValidationType.INTROSPECTION);
        return javax.ws.rs.core.Response.ok().cookie(tenantIdCooke).build();
      }
    } catch (MalformedURLException e) {
      LOG.error(e.getMessage(), e);
    }
    return javax.ws.rs.core.Response.status(javax.ws.rs.core.Response.Status.UNAUTHORIZED).build();
  }

  /**
   * Perform introspection based token validation
   * @param accessToken
   * @return
   */
  private boolean introspect(String accessToken) {
    final String introspectionEndpoint = config.getIntrospectionEndpoint();
    if (introspectionEndpoint != null) {
      LOG.debug("Introspection url: " + introspectionEndpoint);
      HttpUrl introspection = HttpUrl.parse(introspectionEndpoint);
      OkHttpClient client = new OkHttpClient();
      String basic = Credentials.basic(clientId, clientSecret);
      RequestBody formBody = new FormBody.Builder().add("token", accessToken).build();
      Request req = new Request.Builder().url(introspection).header(AUTHORIZATION_HEADER, basic).post(formBody).build();
      final Response response;
      try {
        response = client.newCall(req).execute();
        if (response.isSuccessful()) {
          return true;
        }else{
          LOG.error(response.code() + " - " + response.body().string());
        }
      } catch (IOException e) {
        e.printStackTrace();
      }
    }
    return false;
  }

  /**
   * Gets well known configuration of issuer
   * @return
   */
  private static OpenIdConfiguration getOpenIdConfig() {
    if (config != null) {
      return config;
    }
    Map<String, Object> attributes;
    try {
      HttpUrl httpUrl = HttpUrl.parse(Main.opts.issuer).newBuilder().addPathSegments(".well-known/openid-configuration").build();
      final URLConnection urlConnection = httpUrl.url().openConnection();
      urlConnection.setConnectTimeout(5000);
      urlConnection.setReadTimeout(5000);
      final InputStream inputStream = urlConnection.getInputStream();
      final JsonFactory factory = new JsonFactory();
      final JsonParser parser = factory.createParser(inputStream);
      final TypeReference<Map<String, Object>> typeReference = new TypeReference<Map<String, Object>>() {
      };
      attributes = new ObjectMapper().reader().readValue(parser, typeReference);
    } catch (IOException e) {
      LOG.error("Can't locate openId configuration", e);
      attributes = new HashMap<>();
      final HttpUrl base = HttpUrl.parse(Main.opts.issuer);
      attributes.put("authorization_endpoint", base.newBuilder().addPathSegments("oauth2/v1/authorize").build().url());
      attributes.put("token_endpoint", base.newBuilder().addPathSegments("oauth2/v1/token").build().url());
      attributes.put("introspection_endpoint", base.newBuilder().addPathSegments("oauth2/v1/introspect").build().url());
      attributes.put("issuer", base.url());
    }
    config = OpenIdConfiguration.from(attributes);
    return config;
  }
}

/**
 * Object representing well known configuration
 */
class OpenIdConfiguration {
  String issuer;
  String authorizationEndpoint;
  String tokenEndpoint;
  String jwksUri;
  String introspectionEndpoint;
  ValidationType validationType = ValidationType.PUBLIC_KEY;


  public ValidationType getValidationType() {
    return validationType;
  }

  public void setValidationType(ValidationType validationType) {
    this.validationType = validationType;
  }

  public String getIntrospectionEndpoint() {
    return introspectionEndpoint;
  }

  public void setIntrospectionEndpoint(String introspectionEndpoint) {
    this.introspectionEndpoint = introspectionEndpoint;
  }

  public static OpenIdConfiguration from(Map<String, Object> map) {
    OpenIdConfiguration config = new OpenIdConfiguration();
    config.setAuthorizationEndpoint(map.get("authorization_endpoint").toString());
    config.setTokenEndpoint(map.get("token_endpoint").toString());
    final Object introspect = map.get("introspection_endpoint");
    config.setIntrospectionEndpoint(introspect != null ? introspect.toString() : null);
    config.setIssuer(map.get("issuer").toString());
    final Object jwks_uri = map.get("jwks_uri");
    config.setJwksUri(jwks_uri != null ? jwks_uri.toString() : null);
    if (jwks_uri == null) {
      config.setValidationType(ValidationType.INTROSPECTION);
    }
    return config;
  }

  public String getIssuer() {
    return issuer;
  }

  public void setIssuer(String issuer) {
    this.issuer = issuer;
  }

  public String getAuthorizationEndpoint() {
    return authorizationEndpoint;
  }

  public void setAuthorizationEndpoint(String authorizationEndpoint) {
    this.authorizationEndpoint = authorizationEndpoint;
  }

  public String getTokenEndpoint() {
    return tokenEndpoint;
  }

  public void setTokenEndpoint(String tokenEndpoint) {
    this.tokenEndpoint = tokenEndpoint;
  }

  public String getJwksUri() {
    return jwksUri;
  }

  public void setJwksUri(String jwksUri) {
    this.jwksUri = jwksUri;
  }

  public enum ValidationType{
    PUBLIC_KEY,
    INTROSPECTION
  }

  public enum DomainType{
    IPv6,
    IPv4,
    FQDN,
    SHORT_NAME
  }
}
