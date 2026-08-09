export namespace main {
	
	export class ApresentacaoDTO {
	    Caminho: string;
	    Tipo: string;
	    HTML: string;
	    PDFBase64: string;
	    Erro: string;
	
	    static createFrom(source: any = {}) {
	        return new ApresentacaoDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Caminho = source["Caminho"];
	        this.Tipo = source["Tipo"];
	        this.HTML = source["HTML"];
	        this.PDFBase64 = source["PDFBase64"];
	        this.Erro = source["Erro"];
	    }
	}
	export class AssinaturaDTO {
	    Nome: string;
	    Base64: string;
	    Ativa: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AssinaturaDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Nome = source["Nome"];
	        this.Base64 = source["Base64"];
	        this.Ativa = source["Ativa"];
	    }
	}
	export class AtualizacaoDTO {
	    Disponivel: boolean;
	    Versao: string;
	    Notas: string;
	    Erro: string;
	
	    static createFrom(source: any = {}) {
	        return new AtualizacaoDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Disponivel = source["Disponivel"];
	        this.Versao = source["Versao"];
	        this.Notas = source["Notas"];
	        this.Erro = source["Erro"];
	    }
	}
	export class RegistroDTO {
	    Arquivo: string;
	    Codigo: string;
	    DataReferencia: string;
	    GanhoMesReais: number;
	    GanhoAnoReais: number;
	    RentMesPct: number;
	    RentAnoPct: number;
	    CDIMesPct: number;
	    CDIAnoPct: number;
	    Ganho12MReais: number;
	    Rent12MPct: number;
	    CDI12MPct: number;
	    Patrimonio: number;
	    Copiado: boolean;
	    RentFmt: string;
	    RentAFmt: string;
	    Rent12MFmt: string;
	    PercFmt: string;
	    PercAFmt: string;
	    Perc12MFmt: string;
	    CDIFmt: string;
	    CDIAFmt: string;
	    CDI12MFmt: string;
	
	    static createFrom(source: any = {}) {
	        return new RegistroDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Arquivo = source["Arquivo"];
	        this.Codigo = source["Codigo"];
	        this.DataReferencia = source["DataReferencia"];
	        this.GanhoMesReais = source["GanhoMesReais"];
	        this.GanhoAnoReais = source["GanhoAnoReais"];
	        this.RentMesPct = source["RentMesPct"];
	        this.RentAnoPct = source["RentAnoPct"];
	        this.CDIMesPct = source["CDIMesPct"];
	        this.CDIAnoPct = source["CDIAnoPct"];
	        this.Ganho12MReais = source["Ganho12MReais"];
	        this.Rent12MPct = source["Rent12MPct"];
	        this.CDI12MPct = source["CDI12MPct"];
	        this.Patrimonio = source["Patrimonio"];
	        this.Copiado = source["Copiado"];
	        this.RentFmt = source["RentFmt"];
	        this.RentAFmt = source["RentAFmt"];
	        this.Rent12MFmt = source["Rent12MFmt"];
	        this.PercFmt = source["PercFmt"];
	        this.PercAFmt = source["PercAFmt"];
	        this.Perc12MFmt = source["Perc12MFmt"];
	        this.CDIFmt = source["CDIFmt"];
	        this.CDIAFmt = source["CDIAFmt"];
	        this.CDI12MFmt = source["CDI12MFmt"];
	    }
	}
	export class ClienteRentabilidadeDTO {
	    Codigo: string;
	    Nome: string;
	    Registro?: RegistroDTO;
	    FestasEnviado: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ClienteRentabilidadeDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Codigo = source["Codigo"];
	        this.Nome = source["Nome"];
	        this.Registro = this.convertValues(source["Registro"], RegistroDTO);
	        this.FestasEnviado = source["FestasEnviado"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BaseClientesDTO {
	    ClientDB: Record<string, string>;
	    ClientEmails: Record<string, string>;
	    Clientes: ClienteRentabilidadeDTO[];
	
	    static createFrom(source: any = {}) {
	        return new BaseClientesDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ClientDB = source["ClientDB"];
	        this.ClientEmails = source["ClientEmails"];
	        this.Clientes = this.convertValues(source["Clientes"], ClienteRentabilidadeDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CampoDTO {
	    Key: string;
	    Label: string;
	    Placeholder: string;
	    Type: string;
	
	    static createFrom(source: any = {}) {
	        return new CampoDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Key = source["Key"];
	        this.Label = source["Label"];
	        this.Placeholder = source["Placeholder"];
	        this.Type = source["Type"];
	    }
	}
	export class CategoriaDTO {
	    Group: string;
	    Label: string;
	    IntroFrase: string;
	    Anexo: string;
	    SoPadronizado: boolean;
	    OperacaoUnica: boolean;
	    Fields: CampoDTO[];
	
	    static createFrom(source: any = {}) {
	        return new CategoriaDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Group = source["Group"];
	        this.Label = source["Label"];
	        this.IntroFrase = source["IntroFrase"];
	        this.Anexo = source["Anexo"];
	        this.SoPadronizado = source["SoPadronizado"];
	        this.OperacaoUnica = source["OperacaoUnica"];
	        this.Fields = this.convertValues(source["Fields"], CampoDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CatalogoEmailDTO {
	    Produtos: string[];
	    Categorias: CategoriaDTO[];
	    InfoEstruturadas: string;
	
	    static createFrom(source: any = {}) {
	        return new CatalogoEmailDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Produtos = source["Produtos"];
	        this.Categorias = this.convertValues(source["Categorias"], CategoriaDTO);
	        this.InfoEstruturadas = source["InfoEstruturadas"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class FalhaDTO {
	    Arquivo: string;
	    Erro: string;
	
	    static createFrom(source: any = {}) {
	        return new FalhaDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Arquivo = source["Arquivo"];
	        this.Erro = source["Erro"];
	    }
	}
	export class InfoCadeiaDTO {
	    AtualizadoEm: string;
	    Origem: string;
	    NumCertificados: number;
	    Erro: string;
	
	    static createFrom(source: any = {}) {
	        return new InfoCadeiaDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.AtualizadoEm = source["AtualizadoEm"];
	        this.Origem = source["Origem"];
	        this.NumCertificados = source["NumCertificados"];
	        this.Erro = source["Erro"];
	    }
	}
	export class PreferenciasDTO {
	    Tema: string;
	    Acento: string;
	    ModoEmail: string;
	    TabelaPrevidenciaria: string;
	    Visao: string;
	    Fonte: string;
	    ModoApresentacao: boolean;
	    ModoFestas: boolean;
	    EmailRemetente: string;
	    AssessorNome: string;
	    AssessorEmail: string;
	    OrdemNav: string[];
	    OrdemNavOcultos: string[];
	    TemApresentacao: boolean;
	    RecortePersonalizado: boolean;
	    RecorteX0: number;
	    RecorteY0: number;
	    RecorteX1: number;
	    RecorteY1: number;
	    RecortePadraoX0: number;
	    RecortePadraoY0: number;
	    RecortePadraoX1: number;
	    RecortePadraoY1: number;
	
	    static createFrom(source: any = {}) {
	        return new PreferenciasDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Tema = source["Tema"];
	        this.Acento = source["Acento"];
	        this.ModoEmail = source["ModoEmail"];
	        this.TabelaPrevidenciaria = source["TabelaPrevidenciaria"];
	        this.Visao = source["Visao"];
	        this.Fonte = source["Fonte"];
	        this.ModoApresentacao = source["ModoApresentacao"];
	        this.ModoFestas = source["ModoFestas"];
	        this.EmailRemetente = source["EmailRemetente"];
	        this.AssessorNome = source["AssessorNome"];
	        this.AssessorEmail = source["AssessorEmail"];
	        this.OrdemNav = source["OrdemNav"];
	        this.OrdemNavOcultos = source["OrdemNavOcultos"];
	        this.TemApresentacao = source["TemApresentacao"];
	        this.RecortePersonalizado = source["RecortePersonalizado"];
	        this.RecorteX0 = source["RecorteX0"];
	        this.RecorteY0 = source["RecorteY0"];
	        this.RecorteX1 = source["RecorteX1"];
	        this.RecorteY1 = source["RecorteY1"];
	        this.RecortePadraoX0 = source["RecortePadraoX0"];
	        this.RecortePadraoY0 = source["RecortePadraoY0"];
	        this.RecortePadraoX1 = source["RecortePadraoX1"];
	        this.RecortePadraoY1 = source["RecortePadraoY1"];
	    }
	}
	export class InicioDTO {
	    TemPasta: boolean;
	    Pasta: string;
	    Modelo: string;
	    ModeloFestas: string;
	    ClientDB: Record<string, string>;
	    ClientEmails: Record<string, string>;
	    Prefs: PreferenciasDTO;
	
	    static createFrom(source: any = {}) {
	        return new InicioDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TemPasta = source["TemPasta"];
	        this.Pasta = source["Pasta"];
	        this.Modelo = source["Modelo"];
	        this.ModeloFestas = source["ModeloFestas"];
	        this.ClientDB = source["ClientDB"];
	        this.ClientEmails = source["ClientEmails"];
	        this.Prefs = this.convertValues(source["Prefs"], PreferenciasDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ItemEmailEntrada {
	    Group: string;
	    Label: string;
	    Valores: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new ItemEmailEntrada(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Group = source["Group"];
	        this.Label = source["Label"];
	        this.Valores = source["Valores"];
	    }
	}
	export class ItemRespostaTypeformDTO {
	    Pergunta: string;
	    Valor: string;
	
	    static createFrom(source: any = {}) {
	        return new ItemRespostaTypeformDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Pergunta = source["Pergunta"];
	        this.Valor = source["Valor"];
	    }
	}
	export class PaginaPDFEditadaDTO {
	    JPEGBase64: string;
	    LarguraPt: number;
	    AlturaPt: number;
	
	    static createFrom(source: any = {}) {
	        return new PaginaPDFEditadaDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.JPEGBase64 = source["JPEGBase64"];
	        this.LarguraPt = source["LarguraPt"];
	        this.AlturaPt = source["AlturaPt"];
	    }
	}
	export class PastaDTO {
	    Pasta: string;
	    Modelo: string;
	    ModeloFestas: string;
	
	    static createFrom(source: any = {}) {
	        return new PastaDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Pasta = source["Pasta"];
	        this.Modelo = source["Modelo"];
	        this.ModeloFestas = source["ModeloFestas"];
	    }
	}
	
	export class ProcessamentoDTO {
	    Sucesso: number;
	    Falhas: FalhaDTO[];
	    Clientes: ClienteRentabilidadeDTO[];
	
	    static createFrom(source: any = {}) {
	        return new ProcessamentoDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Sucesso = source["Sucesso"];
	        this.Falhas = this.convertValues(source["Falhas"], FalhaDTO);
	        this.Clientes = this.convertValues(source["Clientes"], ClienteRentabilidadeDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class VerificacaoDTO {
	    Nome: string;
	    Passou: boolean;
	    Detalhe: string;
	
	    static createFrom(source: any = {}) {
	        return new VerificacaoDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Nome = source["Nome"];
	        this.Passou = source["Passou"];
	        this.Detalhe = source["Detalhe"];
	    }
	}
	export class ResultadoValidacaoDTO {
	    Estado: string;
	    Motivo: string;
	    Formato: string;
	    NomeSignatario: string;
	    CPF: string;
	    CNPJ: string;
	    ACEmissora: string;
	    DataAssinatura: string;
	    TemCarimboTempo: boolean;
	    Verificacoes: VerificacaoDTO[];
	    TempoProcessamentoMs: number;
	
	    static createFrom(source: any = {}) {
	        return new ResultadoValidacaoDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Estado = source["Estado"];
	        this.Motivo = source["Motivo"];
	        this.Formato = source["Formato"];
	        this.NomeSignatario = source["NomeSignatario"];
	        this.CPF = source["CPF"];
	        this.CNPJ = source["CNPJ"];
	        this.ACEmissora = source["ACEmissora"];
	        this.DataAssinatura = source["DataAssinatura"];
	        this.TemCarimboTempo = source["TemCarimboTempo"];
	        this.Verificacoes = this.convertValues(source["Verificacoes"], VerificacaoDTO);
	        this.TempoProcessamentoMs = source["TempoProcessamentoMs"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

