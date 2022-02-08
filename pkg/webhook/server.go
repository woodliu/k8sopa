package webhook

import (
    "context"
    "crypto/tls"
    "encoding/json"
    "flag"
    "fmt"
    frameworkClient "github.com/open-policy-agent/frameworks/constraint/pkg/client"
    "github.com/woodliu/k8sopa/pkg/register"
    "io/ioutil"
    v1 "k8s.io/api/admission/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "net/http"
    "stash.weimob.com/devops/go_common/log"
)
type WhSvrParameters struct {
    port           int    // webhook server port
    certFile       string // path to the x509 certificate for https
    keyFile        string // path to the x509 private key matching `CertFile`
}

type Server struct {
    server *http.Server
    k8sClientSet    *kubernetes.Clientset
    frameworkClient *frameworkClient.Client
}
func (svr *Server)Stop(){
    svr.server.Shutdown(context.Background())
}

func Start()*Server{
    var parameters WhSvrParameters

    // get command line parameters
    flag.IntVar(&parameters.port, "port", 443, "Webhook server port.")
    flag.StringVar(&parameters.certFile, "tlsCertFile", "/etc/webhook/certs/cert.pem", "File containing the x509 Certificate for HTTPS.")
    flag.StringVar(&parameters.keyFile, "tlsKeyFile", "/etc/webhook/certs/key.pem", "File containing the x509 private key to --tlsCertFile.")
    flag.Parse()

    pair, err := tls.LoadX509KeyPair(parameters.certFile, parameters.keyFile)
    if err != nil {
        log.Errorf("Failed to load key pair: %v", err)
    }

    svr := &Server{
        server: &http.Server{
            Addr:      fmt.Sprintf(":%v", parameters.port),
            TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
        },
    }

    // define http server and server handler
    mux := http.NewServeMux()
    mux.HandleFunc("/validate", svr.serve)
    svr.server.Handler = mux

    // start webhook server in new routine
    go func() {
        if err := svr.server.ListenAndServeTLS("", ""); err != nil {
            log.Errorf("Failed to listen and serve webhook server: %v", err)
        }
    }()

    log.Info("webhook server started")
    return svr
}

func (svr *Server) serve(w http.ResponseWriter, r *http.Request) {
    var body []byte
    if r.Body != nil {
        if data, err := ioutil.ReadAll(r.Body); err == nil {
            body = data
        }
    }
    if len(body) == 0 {
        log.Errorf("empty body")
        http.Error(w, "empty body", http.StatusBadRequest)
        return
    }

    // verify the content type is accurate
    contentType := r.Header.Get("Content-Type")
    if contentType != "application/json" {
        log.Errorf("Content-Type=%s, expect application/json", contentType)
        http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
        return
    }

    var admissionResponse *v1.AdmissionResponse
    ar := v1.AdmissionReview{}
    if _, _, err := register.Deserializer.Decode(body, nil, &ar); err != nil {
        log.Errorf("Can't decode body: %v", err)
        admissionResponse = &v1.AdmissionResponse{
            Result: &metav1.Status{
                Message: err.Error(),
            },
        }
    } else {
        if r.URL.Path == "/validate" {
            resp := svr.validate(&ar)
            admissionResponse = &resp.AdmissionResponse
        }
    }

    admissionReview := v1.AdmissionReview{}
    if admissionResponse != nil {
        admissionReview.Response = admissionResponse
        if ar.Request != nil {
            admissionReview.Response.UID = ar.Request.UID
        }
    }

    resp, err := json.Marshal(admissionReview)
    if err != nil {
        log.Errorf("Can't encode response: %v", err)
        http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
    }
    log.Info("Ready to write reponse ...")
    if _, err := w.Write(resp); err != nil {
        log.Errorf("Can't write response: %v", err)
        http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
    }
}